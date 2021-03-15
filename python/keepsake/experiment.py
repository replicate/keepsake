try:
    # backport is incompatible with 3.7+, so we must use built-in
    import dataclasses
    from dataclasses import InitVar, dataclass, field
except ImportError:
    from ._vendor.dataclasses import dataclass, InitVar, field
    from ._vendor import dataclasses  # type: ignore

import datetime
import getpass
import html
import json
import math
import os
import shlex
import sys
from typing import (
    TYPE_CHECKING,
    Any,
    Callable,
    Dict,
    List,
    MutableSequence,
    Optional,
    Tuple,
)

from . import console
from .checkpoint import Checkpoint, CheckpointList, PrimaryMetric
from .metadata import parse_rfc3339, rfc3339_datetime
from .packages import get_imported_packages
from .system import get_python_version
from .validate import check_path
from .version import version

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
    user: Optional[str] = None
    host: Optional[str] = None
    command: Optional[str] = None
    config: Optional[dict] = None
    path: Optional[str] = None
    params: Optional[Dict[str, Any]] = None
    python_version: Optional[str] = None
    python_packages: Optional[Dict[str, str]] = None
    keepsake_version: Optional[str] = None
    checkpoints: CheckpointList = field(default_factory=CheckpointList)

    def __post_init__(self, project: "Project"):
        self._project = project
        self._step = -1

    def short_id(self):
        return self.id[:7]

    def _validate(self) -> List[str]:
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
        # protobuf 3 doesn't have optionals, so path=None becomes ""
        # and we have no way of differentiating between empty strings
        # and Nones
        if path == "":
            raise ValueError(
                "path cannot be an empty string. Please use path=None or omit path if you don't want to save any files."
            )

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

        # Auto-increment step if not provided
        if step is None:
            step = self._step + 1
        # Remember the current step
        self._step = step

        checkpoint = self._project._daemon().create_checkpoint(
            experiment=self,
            path=path,
            step=step,
            metrics=metrics,
            primary_metric=primary_metric_dict,
            quiet=quiet,
        )
        self.checkpoints.append(checkpoint)
        self._save(quiet=quiet)
        return checkpoint

    def _save(self, quiet: bool):
        """
        Save this experiment's metadata to repository.
        """
        self._project._daemon().save_experiment(self, quiet=quiet)
        return

    def refresh(self):
        """
        Update this experiment with the latest data from the repository.
        """
        exp = self._project._daemon().get_experiment(experiment_id_prefix=self.id)

        for field in dataclasses.fields(exp):
            if field.name != "project":
                value = getattr(exp, field.name)
                setattr(self, field.name, value)
        for chk in self.checkpoints:
            chk._experiment = self

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
            "python_version": self.python_version,
            "python_packages": self.python_packages,
            "checkpoints": [c.to_json() for c in self.checkpoints],
            "keepsake_version": version,
        }

    def stop(self):
        """
        Stop an experiment.

        Experiments running in a script will eventually timeout, but when running in a notebook,
        you are required to call this method to mark an experiment as stopped.
        """
        self._project._daemon().stop_experiment(self.id)

    def delete(self):
        """
        Delete this experiment and all associated checkpoints.
        """
        # We should consolidate delete logic, see https://github.com/replicate/keepsake/issues/332
        # It's also slow https://github.com/replicate/keepsake/issues/333
        self._project._daemon().delete_experiment(self.id)

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
        valid_metric = lambda m: m is not None and not math.isnan(m)
        primary_metric_checkpoints = [
            chk
            for chk in self.checkpoints
            if chk.primary_metric
            and valid_metric(chk.metrics[chk.primary_metric["name"]])
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
        return self._project._daemon().experiment_is_running(self.id)

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
        return self.checkpoints.plot(metric, logy, plot_only)

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
        # We should add status here, see https://github.com/replicate/keepsake/issues/334
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
        command = " ".join(map(shlex.quote, sys.argv))
        return self.project._daemon().create_experiment(
            path=path,
            params=params,
            command=command,
            python_version=get_python_version(),
            python_packages=get_imported_packages(),
            quiet=quiet,
            disable_hearbeat=disable_heartbeat,
        )

    def get(self, experiment_id_prefix) -> Experiment:
        """
        Returns the experiment with the given ID.
        """
        return self.project._daemon().get_experiment(
            experiment_id_prefix=experiment_id_prefix,
        )

    def list(self, filter: Optional[Callable[[Any], bool]] = None) -> "ExperimentList":
        """
        Return all experiments for a project, sorted by creation date.
        """
        experiments = self.project._daemon().list_experiments()
        result: ExperimentList = ExperimentList()
        for exp in experiments:
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
        plotted_label = plt.gca().get_ylabel() or metric

        if metric != plotted_label:
            plt.figure()

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
