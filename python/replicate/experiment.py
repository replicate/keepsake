try:
    # backport is incompatible with 3.7+, so we must use built-in
    from dataclasses import dataclass, InitVar, field
except ImportError:
    from ._vendor.dataclasses import dataclass, InitVar, field
import getpass
import os
import html
import datetime
import json
import shlex
import sys
from typing import (
    Dict,
    Any,
    Optional,
    Tuple,
    List,
    TYPE_CHECKING,
    MutableSequence,
    Callable,
)

from . import console
from .exceptions import DoesNotExistError, NewerRepositoryVersion
from .checkpoint import (
    Checkpoint,
    PrimaryMetric,
    CheckpointList,
)
from .hash import random_hash
from .heartbeat import Heartbeat, DEFAULT_REFRESH_INTERVAL
from .json import CustomJSONEncoder
from .metadata import rfc3339_datetime, parse_rfc3339
from .packages import get_imported_packages
from .validate import check_path
from .version import version
from .constants import (
    REPOSITORY_VERSION,
    PYTHON_REFERENCE_DOCS_URL,
    HEARTBEAT_MISS_TOLERANCE,
    EXPERIMENT_STATUS_RUNNING,
    EXPERIMENT_STATUS_STOPPED,
)

if TYPE_CHECKING:
    from .project import Project


@dataclass
class Experiment:
    """
    A run of a training script.
    """

    project: InitVar["Project"]

    id: str
    created: datetime.datetime
    user: str
    host: str
    command: str
    config: dict
    status: str
    path: Optional[str] = None
    params: Optional[Dict[str, Any]] = None
    python_packages: Optional[Dict[str, str]] = None
    replicate_version: Optional[str] = None
    checkpoints: CheckpointList = field(default_factory=CheckpointList)

    def __post_init__(self, project: "Project"):
        self._project = project
        self._heartbeat = None

    def short_id(self):
        return self.id[:7]

    def validate(self) -> List[str]:
        errors = []

        if self.params is not None:
            if isinstance(self.params, dict):
                for key, value in self.params.items():
                    try:
                        json.dumps(value)
                    except (ValueError, TypeError, OverflowError):
                        errors.append(
                            "Failed to serialize the param '{}' to JSON. Make sure it's only using basic types (str, int, float, bool, dict, list, None)".format(
                                key
                            )
                        )
            else:
                errors.append("params must be a dictionary")

        if self.path is not None:
            errors.extend(check_path("experiment", self.path))

        return errors

    @console.catch_and_print_exceptions(msg="Error creating checkpoint")
    def checkpoint(
        self,
        path: Optional[str] = None,
        step: Optional[int] = None,
        metrics: Optional[Dict[str, Any]] = None,
        primary_metric: Optional[Tuple[str, str]] = None,
        quiet: bool = False,
    ) -> Optional[Checkpoint]:
        """
        Create a checkpoint within this experiment.

        This saves the metrics at this point, and makes a copy of the file or directory passed to `path`, which could be weights or any other artifact.
        """
        # TODO(bfirsh): display warning if primary_metric changes in an experiment
        # FIXME: store as tuple throughout for consistency?
        primary_metric_dict: Optional[PrimaryMetric] = None
        if primary_metric is not None:
            if len(primary_metric) != 2:
                # Not having a primary_metric isn't a fatal problem
                console.error(
                    "Not setting primary_metric on checkpoint: it must be a tuple of (name, goal), where name corresponds to a metric key, and goal is either 'maximize' or 'minimize'"
                )
            else:
                primary_metric_dict = {
                    "name": primary_metric[0],
                    "goal": primary_metric[1],
                }

        checkpoint = Checkpoint(
            id=random_hash(),
            created=datetime.datetime.utcnow(),
            path=path,
            step=step,
            metrics=metrics,
            primary_metric=primary_metric_dict,
        )
        if not quiet:
            if path is None:
                console.info("Creating checkpoint {}".format(checkpoint.short_id()))
            else:
                console.info(
                    "Creating checkpoint {}: copying '{}' to '{}'...".format(
                        checkpoint.short_id(),
                        checkpoint.path,
                        self._project._get_repository().root_url(),
                    )
                )

        errors = checkpoint.validate()
        if errors:
            for error in errors:
                console.error("Not saving checkpoint: " + error)
            return checkpoint

        checkpoint._experiment = self

        # Upload files before writing metadata so if it is cancelled, there isn't metadata pointing at non-existent data
        if checkpoint.path is not None:
            tar_path = checkpoint._repository_tar_path()
            repository = self._project._get_repository()
            repository.put_path_tar(self._project.directory, tar_path, checkpoint.path)

        self.checkpoints.append(checkpoint)
        self.save()

        if self._heartbeat is not None:
            self._heartbeat.ensure_running()

        return checkpoint

    def save(self):
        """
        Save this experiment's metadata to repository.
        """
        repository = self._project._get_repository()
        repository.put(
            self._metadata_path(),
            json.dumps(self.to_json(), indent=2, cls=CustomJSONEncoder),
        )

    @classmethod
    def from_json(cls, project: "Project", data: Dict[str, Any]) -> "Experiment":
        data = data.copy()
        data["created"] = parse_rfc3339(data["created"])
        data["checkpoints"] = CheckpointList(
            [Checkpoint.from_json(d) for d in data.get("checkpoints", [])]
        )
        experiment = Experiment(project=project, **data)
        experiment.status = (
            EXPERIMENT_STATUS_RUNNING
            if experiment.is_running()
            else EXPERIMENT_STATUS_STOPPED
        )
        for chk in experiment.checkpoints:
            chk._experiment = experiment
        return experiment

    def to_json(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "created": rfc3339_datetime(self.created),
            "params": self.params,
            "user": self.user,
            "host": self.host,
            "command": self.command,
            "config": self.config,
            "status": EXPERIMENT_STATUS_RUNNING
            if self.is_running()
            else EXPERIMENT_STATUS_STOPPED,
            "path": self.path,
            "python_packages": self.python_packages,
            "checkpoints": [c.to_json() for c in self.checkpoints],
            "replicate_version": version,
        }

    def start_heartbeat(self):
        self._heartbeat = Heartbeat(
            experiment_id=self.id,
            repository_url=self._project._get_config()["repository"],
            path=self._heartbeat_path(),
        )
        self._heartbeat.start()

    def stop(self):
        """
        Stop an experiment.

        Experiments running in a script will eventually timeout, but when running in a notebook,
        you are required to call this method to mark an experiment as stopped.
        """
        if self._heartbeat is not None:
            self._heartbeat.kill()
            self._heartbeat = None
        self._project._get_repository().delete(self._heartbeat_path())

    def delete(self):
        """
        Delete this experiment and all associated checkpoints.
        """
        # We should consolidate delete logic, see https://github.com/replicate/replicate/issues/332
        # It's also slow https://github.com/replicate/replicate/issues/333
        repository = self._project._get_repository()
        console.info(
            "Deleting {} checkpoints in experiment {}".format(
                len(self.checkpoints), self.short_id()
            )
        )
        for checkpoint in self.checkpoints:
            repository.delete(checkpoint._repository_tar_path())
        console.info("Deleting experiment: {}".format(self.short_id()))
        repository.delete(self._heartbeat_path())
        repository.delete(self._repository_tar_path())
        repository.delete(self._metadata_path())

    def latest(self) -> Optional[Checkpoint]:
        """
        Get the latest checkpoint in this experiment, or None
        if there are no checkpoints.
        """
        if self.checkpoints:
            return self.checkpoints[-1]
        return None

    def best(self) -> Optional[Checkpoint]:
        """
        Get the best checkpoint in this experiment, or None
        if there are no checkpoints or no checkpoint has a primary
        metric.
        """
        if not self.checkpoints:
            return None
        primary_metric_checkpoints = [
            chk for chk in self.checkpoints if chk.primary_metric
        ]
        if not primary_metric_checkpoints:
            return None
        name = primary_metric_checkpoints[0].primary_metric["name"]  # type: ignore
        goal = primary_metric_checkpoints[0].primary_metric["goal"]  # type: ignore
        if not all(
            chk.primary_metric["name"] == name for chk in primary_metric_checkpoints  # type: ignore
        ):
            console.warn(
                "Not all checkpoints in experiment {} have the same primary metric name".format(
                    self.short_id()
                )
            )
        if not all(
            chk.primary_metric["goal"] == goal for chk in primary_metric_checkpoints  # type: ignore
        ):
            console.warn(
                "Not all checkpoints in experiment {} have the same primary metric goal".format(
                    self.short_id()
                )
            )

        if goal == "minimize":
            key = lambda chk: -chk.metrics[name]
        else:
            key = lambda chk: chk.metrics[name]

        return sorted(primary_metric_checkpoints, key=key)[-1]

    def is_running(self) -> bool:
        """
        Check whether the experiment is running or not.
        If the last heartbeat recorded is greater than the last tolerable
        heartbeat then return False else True
        In case the heartbeat metadata file is not present which means the
        experiment was stopped the function returns False.
        """
        try:
            repository = self._project._get_repository()
            heartbeat_metadata_bytes = repository.get(self._heartbeat_path())
            heartbeat_metadata = json.loads(heartbeat_metadata_bytes)
        except Exception as e:
            return False
        now = datetime.datetime.utcnow()
        last_heartbeat = parse_rfc3339(heartbeat_metadata["last_heartbeat"])
        last_tolerable_heartbeat = (
            now - DEFAULT_REFRESH_INTERVAL * HEARTBEAT_MISS_TOLERANCE
        )
        return last_tolerable_heartbeat < last_heartbeat

    def _heartbeat_path(self) -> str:
        return "metadata/heartbeats/{}.json".format(self.id)

    def _repository_tar_path(self) -> str:
        return "experiments/{}.tar.gz".format(self.id)

    def _metadata_path(self) -> str:
        return "metadata/experiments/{}.json".format(self.id)

    def primary_metric(self) -> str:
        """
        Get the shared primary metric for the list of checkpoints in
        this experiment. If no shared primary metric exists, raises
        ValueError.
        """
        return self.checkpoints.primary_metric()

    def plot(self, metric: Optional[str] = None, logy=False, plot_only=False):
        """
        Plot a metric for all the checkpoints in this experiment. If
        no metric is specified, defaults to the shared primary metric.
        """
        self.checkpoints.plot(metric, logy, plot_only)

    @property
    def duration(self) -> Optional[datetime.timedelta]:
        if not self.checkpoints:
            return None

        last = self.checkpoints[-1]
        return last.created - self.created

    def _repr_html_(self) -> str:
        out = '<p><b><pre style="display: inline">Experiment(id="{}")</pre></b></p>'.format(
            self.id
        )
        out += "<p>"
        # We should add status here, see https://github.com/replicate/replicate/issues/334
        for field in ["created", "host", "user", "command", "duration"]:
            out += '<pre style="display: inline">{:10s}</pre> {}<br/>'.format(
                html.escape(field) + ":", html.escape(str(getattr(self, field)))
            )
        out += "</p>"
        out += '<p><b><pre style="display: inline">params:</pre></b></p>'
        out += '<table><tr><th style="text-align: left">Name</th><th style="text-align: left">Value</th></tr>'
        if self.params is not None:
            for key, value in self.params.items():
                out += '<tr><td style="text-align: left"><pre>{}</pre></td><td style="text-align: left">{}</td>'.format(
                    html.escape(key), html.escape(str(value))
                )
        out += "</table>"

        out += '<p><b><pre style="display: inline">checkpoints:</pre></b></p>'
        metrics = set()
        for chk in self.checkpoints:
            if chk.metrics is not None:
                metrics |= set(chk.metrics.keys())

        chk_headings = ["short_id", "step", "created"] + [
            'metrics["{}"]'.format(html.escape(str(m))) for m in metrics
        ]
        out += "<table><tr>"
        for heading in chk_headings:
            out += '<th style="text-align: left"><pre>{}</pre></th>'.format(
                html.escape(heading)
            )
        out += "</tr>"
        best = self.best()
        for chk in self.checkpoints:
            out += "<tr>"
            is_best = False
            values = []
            for heading in chk_headings:
                if heading.startswith('metrics["'):
                    name = heading.split('"')[1].split('"')[0]
                    if chk.metrics is None:
                        value = ""
                    else:
                        value = chk.metrics.get(name)
                    if (
                        chk == best
                        and chk.primary_metric
                        and chk.primary_metric["name"] == name
                    ):
                        value = html.escape(str(value)) + " (best)"
                        is_best = True
                elif heading == "short_id":
                    value = chk.short_id()
                else:
                    value = html.escape(str(getattr(chk, heading)))
                values.append(value)

            for value in values:
                out += '<td style="text-align: left">'
                if is_best:
                    out += "<b>{}</b>".format(html.escape(str(value)))
                else:
                    out += html.escape(str(value))
            out += "</tr>"
        out += "</table>"

        return out


@dataclass
class ExperimentCollection:
    """
    An object for managing experiments in a project.
    """

    # ExperimentCollection is initialized on import, so don't do anything slow on init

    project: "Project"

    def create(
        self, path=None, params=None, quiet=False, disable_heartbeat=False
    ) -> Experiment:
        root_url = self.project._get_repository().root_url()

        # check that the project's repository version isn't
        # higher than what this version of replicate can write.
        # projects have to use a single consistent repository version.
        project_spec = self.project._load_project_spec()
        if project_spec is None:
            self.project._write_project_spec(version=REPOSITORY_VERSION)
        elif project_spec.version > REPOSITORY_VERSION:
            raise NewerRepositoryVersion(root_url)

        command = " ".join(map(shlex.quote, sys.argv))
        config = self.project._get_config()
        experiment = Experiment(
            project=self.project,
            id=random_hash(),
            created=datetime.datetime.utcnow(),
            path=path,
            params=params,
            config=config,
            status=EXPERIMENT_STATUS_RUNNING,
            user=os.getenv("REPLICATE_INTERNAL_USER", getpass.getuser()),
            host=os.getenv("REPLICATE_INTERNAL_HOST", ""),
            command=os.getenv("REPLICATE_INTERNAL_COMMAND", command),
            python_packages=get_imported_packages(),
        )

        if not quiet:
            if path is None:
                console.info("Creating experiment {}".format(experiment.short_id()))
            else:
                console.info(
                    "Creating experiment {}: copying '{}' to '{}'...".format(
                        experiment.short_id(), experiment.path, root_url,
                    )
                )

        errors = experiment.validate()
        if errors:
            if len(errors) == 1:
                s = [f"Could not create Replicate experiment: {errors[0]}"]
            else:
                s = ["Could not create Replicate experiment:"]
                for error in errors:
                    s.append(f"- {error}")
            s.append("")
            s.append(f"For help, see the docs: {PYTHON_REFERENCE_DOCS_URL}")
            raise ValueError("\n".join(s))

        # Upload files before writing metadata so if it is cancelled, there isn't metadata pointing at non-existent data
        if experiment.path is not None:
            repository = self.project._get_repository()
            tar_path = experiment._repository_tar_path()
            repository.put_path_tar(self.project.directory, tar_path, experiment.path)

        experiment.save()

        if not disable_heartbeat:
            experiment.start_heartbeat()

        return experiment

    def get(self, experiment_id) -> Experiment:
        """
        Returns the experiment with the given ID.
        """
        repository = self.project._get_repository()
        ids = []
        for path in repository.list("metadata/experiments/"):
            ids.append(os.path.basename(path).split(".")[0])

        matching_ids = list(filter(lambda i: i.startswith(experiment_id), ids))
        if len(matching_ids) == 0:
            raise DoesNotExistError(
                "'{}' does not match any experiment IDs".format(experiment_id)
            )
        elif len(matching_ids) > 1:
            raise DoesNotExistError(
                "'{}' is ambiguous - it matches {} experiment IDs".format(
                    experiment_id, len(matching_ids)
                )
            )

        data = json.loads(
            repository.get("metadata/experiments/{}.json".format(matching_ids[0]))
        )
        return Experiment.from_json(self.project, data)

    def list(self, filter: Optional[Callable[[Any], bool]] = None) -> List[Experiment]:
        """
        Return all experiments for a project, sorted by creation date.
        """
        repository = self.project._get_repository()
        result: ExperimentList = ExperimentList()
        for path in repository.list("metadata/experiments/"):
            data = json.loads(repository.get(path))
            exp = Experiment.from_json(self.project, data)
            if filter is not None:
                include = False
                try:
                    include = filter(exp)
                except Exception as e:  # pylint: disable=broad-except
                    console.warn(
                        "Failed to apply filter to experiment {}: {}".format(
                            exp.short_id(), e
                        )
                    )
                if include:
                    result.append(exp)
            else:
                result.append(exp)
        result.sort(key=lambda e: e.created)
        return result


class ExperimentList(list, MutableSequence[Experiment]):
    def primary_metric(self) -> str:
        """
        Get the shared primary metric for all experiments in this list
        of experiments. If no shared primary metric exists, raises
        ValueError.
        """
        primary_metric = None
        for exp in self:
            for chk in exp.checkpoints:
                pm = chk.primary_metric["name"]
                if pm is None:
                    continue
                if primary_metric is not None and primary_metric != pm:
                    raise ValueError(
                        "The primary metric differs between the checkpoints in these experiments"
                    )
                primary_metric = pm
        if primary_metric is None:
            raise ValueError(
                "No primary metric is defined for the checkpoints in these experiments"
            )

        return primary_metric

    def plot(self, metric: Optional[str] = None, logy=False):
        """
        Plot a metric for all the checkpoints in this list of
        experiments. If no metric is specified, defaults to the
        shared primary metric.
        """
        import matplotlib.pyplot as plt  # type: ignore

        if metric is None:
            metric = self.primary_metric()

        for exp in self:
            exp.plot(metric, plot_only=True)

        plt.legend(bbox_to_anchor=(1, 1))
        plt.xlabel("step")
        plt.ylabel(metric)

        if logy:
            plt.yscale("log")

    def scatter(self, param: str, metric: Optional[str] = None, logx=False, logy=False):
        """
        Plot a metric against a parameter for all experiments in
        this list. If the experiments define primary metric, the
        metric of best checkpoint will be used, otherwise the metric
        will come from the latest checkpoint.
        """
        import matplotlib.pyplot as plt  # type: ignore

        if metric is None:
            metric = self.primary_metric()

        names = []
        for exp in self:
            chk = exp.best()
            if chk is None:
                chk = exp.latest()
            if chk is None:
                console.warn(
                    "Experiment '{}' does not have any checkpoints".format(
                        exp.short_id()
                    )
                )
                continue
            if metric not in chk.metrics:
                console.warn(
                    "Metric '{}' does not exist in checkpoint {} in experiment {}".format(
                        metric, chk.short_id(), exp.short_id()
                    )
                )
                continue
            if param not in exp.params:
                console.warn(
                    "Parameter '{}' does not exist in experiment {}".format(
                        param, exp.short_id()
                    )
                )
                continue
            plt.scatter([exp.params[param]], [chk.metrics[metric]])
            names.append(exp.short_id())

        plt.legend(names, bbox_to_anchor=(1, 1))
        plt.xlabel(param)
        plt.ylabel(metric)

        if logx:
            plt.xscale("log")
        if logy:
            plt.yscale("log")

    def delete(self):
        """
        Delete all experiments in this list of experiments.
        """
        for exp in self:
            exp.delete()

    def _repr_html_(self):
        show_user = False
        user = None
        for experiment in self:
            if user is not None and user != experiment.user:
                show_user = True
                break
            user = experiment.user

        show_host = False
        host = None
        for experiment in self:
            if host is not None and host != experiment.host:
                show_host = True
                break
            host = experiment.host
        if host == "":
            show_host = False

        def format_checkpoint(checkpoint: Optional[Checkpoint]) -> str:
            if not checkpoint:
                return ""
            parens = []
            if checkpoint.step is not None:
                parens.append("step {}".format(checkpoint.step))
            if (
                checkpoint.primary_metric
                and checkpoint.metrics
                and checkpoint.metrics.get(checkpoint.primary_metric["name"])
                is not None
            ):
                parens.append(
                    "{}: {}".format(
                        checkpoint.primary_metric["name"],
                        checkpoint.metrics[checkpoint.primary_metric["name"]],
                    )
                )
            if parens:
                return "{} ({})".format(checkpoint.short_id(), "; ".join(parens))
            return checkpoint.short_id()

        headings = ["id", "created"]
        if show_user:
            headings.append("user")
        if show_host:
            headings.append("host")
        headings += ["params", "latest_checkpoint", "best_checkpoint"]
        out = ["<table>"]
        out.append("<tr>")
        for h in headings:
            out.append("<th>")
            out.append(html.escape(h))
            out.append("</th>")
        out.append("</tr>")
        for experiment in self:
            out.append("<tr>")
            for h in headings:
                if h == "latest_checkpoint":
                    d = format_checkpoint(experiment.latest())
                elif h == "best_checkpoint":
                    d = format_checkpoint(experiment.best())
                else:
                    d = getattr(experiment, h)
                    d = str(d)
                    if h == "id":
                        d = d[:7]
                out.append("<th>")
                out.append(html.escape(d))
                out.append("</th>")
            out.append("</tr>")
        out.append("</table>")
        return "".join(out)

    def __getitem__(self, key):
        if isinstance(key, slice):
            indices = range(*key.indices(len(self)))
            return ExperimentList([self[i] for i in indices])
        return super().__getitem__(key)
