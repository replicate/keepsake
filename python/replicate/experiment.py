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
from .exceptions import DoesNotExistError
from .checkpoint import (
    Checkpoint,
    PrimaryMetric,
)
from .hash import random_hash
from .heartbeat import Heartbeat
from .json import CustomJSONEncoder
from .metadata import rfc3339_datetime, parse_rfc3339
from .packages import get_imported_packages
from .validate import check_path

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
    path: Optional[str]
    params: Optional[Dict[str, Any]] = None
    python_packages: Optional[Dict[str, str]] = None
    checkpoints: List[Checkpoint] = field(default_factory=list)

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
                raise ValueError(
                    "primary_metric must be a tuple of (name, goal), where name corresponds to a metric key, and goal is either 'maximize' or 'minimize'"
                )
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
                        self._project._get_storage().root_url(),
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
            tar_path = checkpoint._storage_tar_path()
            storage = self._project._get_storage()
            storage.put_path_tar(self._project.directory, tar_path, checkpoint.path)

        self.checkpoints.append(checkpoint)
        self.save()

        if self._heartbeat is not None:
            self._heartbeat.ensure_running()

        return checkpoint

    def save(self):
        """
        Save this experiment's metadata to storage.
        """
        storage = self._project._get_storage()
        storage.put(
            self._metadata_path(),
            json.dumps(self.to_json(), indent=2, cls=CustomJSONEncoder),
        )

    @classmethod
    def from_json(cls, project: "Project", data: Dict[str, Any]) -> "Experiment":
        data = data.copy()
        data["created"] = parse_rfc3339(data["created"])
        data["checkpoints"] = [
            Checkpoint.from_json(d) for d in data.get("checkpoints", [])
        ]
        experiment = Experiment(project=project, **data)
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
            "path": self.path,
            "python_packages": self.python_packages,
            "checkpoints": [c.to_json() for c in self.checkpoints],
        }

    def start_heartbeat(self):
        self._heartbeat = Heartbeat(
            experiment_id=self.id,
            storage_url=self._project._get_config()["storage"],
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
        self._project._get_storage().delete(self._heartbeat_path())

    def delete(self):
        """
        Delete this experiment and all associated checkpoints.
        """
        # TODO(andreas): this logic should probably live in go,
        # which could then also be parallelized easily
        storage = self._project._get_storage()
        console.info(
            "Deleting {} checkpoints in experiment {}".format(
                len(self.checkpoints), self.short_id()
            )
        )
        for checkpoint in self.checkpoints:
            storage.delete(checkpoint._storage_tar_path())
        console.info("Deleting experiment: {}".format(self.short_id()))
        storage.delete(self._heartbeat_path())
        storage.delete(self._storage_tar_path())
        storage.delete(self._metadata_path())

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

    def _heartbeat_path(self) -> str:
        return "metadata/heartbeats/{}.json".format(self.id)

    def _storage_tar_path(self) -> str:
        return "experiments/{}.tar.gz".format(self.id)

    def _metadata_path(self) -> str:
        return "metadata/experiments/{}.json".format(self.id)

    def primary_metric(self) -> str:
        """
        Get the shared primary metric for the list of checkpoints in
        this experiment. If no shared primary metric exists, raises
        ValueError.
        """
        primary_metric = None
        for chk in self.checkpoints:
            if chk.primary_metric is None:
                continue

            pm = chk.primary_metric["name"]
            if pm is None:
                continue
            if primary_metric is not None and primary_metric != pm:
                # TODO(andreas): should this be another standard error type?
                raise ValueError(
                    "The primary metric differs between the checkpoints in this experiments"
                )
            primary_metric = pm

        if primary_metric is None:
            raise ValueError(
                "No primary metric is defined for the checkpoints in theis experiment"
            )

        return primary_metric

    def plot(self, metric: Optional[str] = None, logy=False, plot_only=False):
        """
        Plot a metric for all the checkpoints in this experiment. If
        no metric is specified, defaults to the shared primary metric.
        """

        # TODO(andreas): smoothing
        import matplotlib.pyplot as plt  # type: ignore

        if metric is None:
            metric = self.primary_metric()

        data = []
        for chk in self.checkpoints:
            if chk.metrics and metric in chk.metrics:
                # TODO(andreas): handle non-numeric metric
                # TODO(andreas): warn if metric doesn't exist in any experiment
                data.append(chk.metrics[metric])
            else:
                data.append(None)

        every_checkpoint_has_step = True
        steps = []
        for chk in self.checkpoints:
            if chk.step is None:
                every_checkpoint_has_step = False
                break
            steps.append(chk.step)
        if not every_checkpoint_has_step:
            steps = list(range(len(data)))

        plt.plot(steps, data, label=self.short_id())

        if not plot_only:
            plt.legend(bbox_to_anchor=(1, 1))
            plt.xlabel("step")
            plt.ylabel(metric)

            if logy:
                plt.yscale("log")

    @property
    def duration(self) -> Optional[datetime.timedelta]:
        if not self.checkpoints:
            return None

        first = self.checkpoints[0]
        last = self.checkpoints[-1]

        return last.created - first.created

    def _repr_html_(self) -> str:
        out = '<p><b><pre style="display: inline">Experiment(id="{}")</pre></b></p>'.format(
            self.id
        )
        out += "<p>"
        # TODO(andreas): add status
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

    def create(self, path=None, params=None, quiet=False) -> Experiment:
        command = " ".join(map(shlex.quote, sys.argv))
        experiment = Experiment(
            project=self.project,
            id=random_hash(),
            created=datetime.datetime.utcnow(),
            path=path,
            params=params,
            config=self.project._get_config(),
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
                        experiment.short_id(),
                        experiment.path,
                        self.project._get_storage().root_url(),
                    )
                )

        errors = experiment.validate()
        if errors:
            for error in errors:
                console.error("Not saving experiment: " + error)
            return experiment

        # Upload files before writing metadata so if it is cancelled, there isn't metadata pointing at non-existent data
        if experiment.path is not None:
            storage = self.project._get_storage()
            tar_path = experiment._storage_tar_path()
            storage.put_path_tar(self.project.directory, tar_path, experiment.path)

        experiment.save()

        return experiment

    def get(self, experiment_id) -> Experiment:
        """
        Returns the experiment with the given ID.
        """
        storage = self.project._get_storage()
        ids = []
        for path in storage.list("metadata/experiments/"):
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
            storage.get("metadata/experiments/{}.json".format(matching_ids[0]))
        )
        return Experiment.from_json(self.project, data)

    def list(self, filter: Optional[Callable[[Any], bool]] = None) -> List[Experiment]:
        """
        Return all experiments for a project, sorted by creation date.
        """
        storage = self.project._get_storage()
        result: ExperimentList = ExperimentList()
        for path in storage.list("metadata/experiments/"):
            data = json.loads(storage.get(path))
            exp = Experiment.from_json(self.project, data)
            if filter is not None:
                include = False
                try:
                    include = filter(exp)
                except Exception as e:
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
                    # TODO(andreas): should this be another standard error type?
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
        # TODO(andreas): smoothing
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
        for exp in self:
            if user is not None and user != exp.user:
                show_user = True
                break
            user = exp.user

        show_host = False
        host = None
        for exp in self:
            if host is not None and host != exp.host:
                show_host = True
                break
            host = exp.host
        if host == "":
            show_host = False

        def format_checkpoint(chk: Optional[Checkpoint]) -> str:
            if not chk:
                return ""
            parens = []
            if chk.step is not None:
                parens.append("step {}".format(chk.step))
            if (
                chk.primary_metric
                and chk.metrics
                and chk.metrics.get(chk.primary_metric["name"]) is not None
            ):
                parens.append(
                    "{}: {}".format(
                        chk.primary_metric["name"],
                        chk.metrics[chk.primary_metric["name"]],
                    )
                )
            if parens:
                return "{} ({})".format(chk.short_id(), "; ".join(parens))
            return chk.short_id()

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
                    d = format_checkpoint(exp.latest())
                elif h == "best_checkpoint":
                    d = format_checkpoint(exp.best())
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
