try:
    # backport is incompatible with 3.7+, so we must use built-in
    from dataclasses import dataclass
except ImportError:
    from ._vendor.dataclasses import dataclass
import datetime
import os
import io
import tempfile
import json
import sys
import html
from typing import Optional, Dict, Any, List, BinaryIO, TYPE_CHECKING, MutableSequence

if sys.version_info >= (3, 8):
    from typing import TypedDict
else:
    from ._vendor.typing_extensions import TypedDict

from . import console
from .json import CustomJSONEncoder
from .metadata import rfc3339_datetime, parse_rfc3339
from .validate import check_path

if TYPE_CHECKING:
    from .experiment import Experiment


class PrimaryMetric(TypedDict):
    name: str
    goal: str


@dataclass
class Checkpoint(object):
    """
    A checkpoint within an experiment. It represents the metrics and the file or directory specified by `path` at a point in time during the experiment.
    """

    id: str
    created: datetime.datetime
    path: Optional[str] = None
    step: Optional[int] = None
    metrics: Optional[Dict[str, Any]] = None
    primary_metric: Optional[PrimaryMetric] = None

    def __post_init__(self):
        self._experiment: Optional["Experiment"] = None

    def short_id(self) -> str:
        return self.id[:7]

    @classmethod
    def from_json(self, data: Dict[str, Any]) -> "Checkpoint":
        data = data.copy()
        data["created"] = parse_rfc3339(data["created"])
        return Checkpoint(**data)

    def to_json(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "created": rfc3339_datetime(self.created),
            "path": self.path,
            "metrics": self.metrics,
            "primary_metric": self.primary_metric,
            "step": self.step,
        }

    def validate(self) -> List[str]:
        errors = []

        if self.path is not None and not isinstance(self.path, str):
            errors.append("path must be a string")

        if self.step is not None and not isinstance(self.step, int):
            errors.append("step must be an integer")

        if self.metrics is not None:
            if isinstance(self.metrics, dict):
                for key, value in self.metrics.items():
                    try:
                        json.dumps(value, cls=CustomJSONEncoder)
                    except (ValueError, TypeError, OverflowError):
                        errors.append(
                            "Failed to serialize the metric '{}' to JSON. Make sure it's only using basic types (str, int, float, bool, dict, list, None)".format(
                                key
                            )
                        )
            else:
                errors.append("metrics must be a dictionary")

        if (
            self.metrics is not None
            and self.primary_metric is not None
            and self.primary_metric.get("name") is not None
            and self.primary_metric["name"] not in self.metrics
        ):
            errors.append(
                "Primary metric '{}' is not defined in metrics".format(
                    self.primary_metric["name"]
                )
            )

        if (
            self.primary_metric is not None
            and self.primary_metric.get("goal") is not None
            and self.primary_metric["goal"].lower() not in ("maximize", "minimize",)
        ):
            errors.append(
                "Primary metric goal must be either 'maximize' or 'minimize'. "
                "For example: ('loss', 'minimize')"
            )

        if self.path is not None and isinstance(self.path, str):
            errors.extend(check_path("checkpoint", self.path))

        return errors

    # TODO(andreas): this is only checking out checkpoints, which
    # is different from the CLI where both checkpoints and
    # experiments can be checked out with `keepsake checkout`.
    # Perhaps we should support experiment.checkout() as well?
    def checkout(self, output_directory: str, quiet: bool = False):
        """
        Copy files from this checkpoint to the output directory.
        """
        assert self._experiment

        self._experiment._project._daemon().checkout_checkpoint(
            self.id, output_directory, quiet=quiet
        )

    def open(self, path: str) -> BinaryIO:
        """
        Read a single file from this checkpoint into memory.
        Returns a file-like object.
        """
        with tempfile.TemporaryDirectory() as tempdir:
            self.checkout(tempdir, quiet=True)
            with open(os.path.join(tempdir, path), "rb") as f:
                # We shouldn't load entire file into memory, see https://github.com/replicate/keepsake/issues/331
                out_f = io.BytesIO(f.read())
            return out_f

    def _repr_html_(self) -> str:
        out = '<p><b><pre style="display: inline">Checkpoint(id="{}")</pre></b></p>'.format(
            self.id
        )
        out += "<p>"
        for field in ["created", "path", "step"]:
            out += '<pre style="display: inline">{:10s}</pre> {}<br/>'.format(
                html.escape(field) + ":", html.escape(str(getattr(self, field)))
            )
        out += "</p>"

        out += '<p><b><pre style="display: inline">metrics:</pre></b></p>'
        out += '<table><tr><th style="text-align: left">Name</th><th style="text-align: left">Value</th></tr>'
        if self.metrics is not None:
            for key, value in self.metrics.items():
                out += '<tr><td style="text-align: left"><pre>{}</pre></td><td style="text-align: left">{}</td>'.format(
                    html.escape(key), html.escape(str(value))
                )
        out += "</table>"

        return out


class CheckpointList(list, MutableSequence[Checkpoint]):
    def primary_metric(self) -> str:
        """
        Get the shared primary metric for this list of checkpoints.
        If no shared primary metric exists, raises ValueError.
        """
        primary_metric = None
        for chk in self:
            if chk.primary_metric is None:
                continue

            pm = chk.primary_metric["name"]
            if pm is None:
                continue
            if primary_metric is not None and primary_metric != pm:
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
        Plot a metric for this list of checkpoints. If no metric is specified,
        defaults to the shared primary metric.
        """
        import matplotlib.pyplot as plt  # type: ignore

        if metric is None:
            metric = self.primary_metric()

        data = []
        for chk in self:
            if chk.metrics and metric in chk.metrics:
                data.append(chk.metrics[metric])
            else:
                data.append(None)

        every_checkpoint_has_step = True
        steps = []
        for chk in self:
            if chk.step is None:
                every_checkpoint_has_step = False
                break
            steps.append(chk.step)
        if not every_checkpoint_has_step:
            steps = list(range(len(data)))

        label = None
        if len(self) > 0:
            label = self[0]._experiment.short_id()
        plt.plot(steps, data, label=label)

        if not plot_only:
            plt.xlabel("step")
            plt.ylabel(metric)
            plt.legend(bbox_to_anchor=(1, 1))

            if logy:
                plt.yscale("log")

    @property
    def metrics(self):
        # Make this eager instead of lazy? See https://github.com/replicate/keepsake/issues/337
        return CheckpointListMetrics(self)

    @property
    def step(self):
        return [chk.step for chk in self]

    def __getitem__(self, key):
        if isinstance(key, slice):
            indices = range(*key.indices(len(self)))
            return CheckpointList([self[i] for i in indices])
        return super().__getitem__(key)


class CheckpointListMetrics:
    def __init__(self, checkpoint_list: CheckpointList):
        self.checkpoint_list = checkpoint_list

    def __getitem__(self, name: str) -> List[Any]:
        values = [
            chk.metrics.get(name) if chk.metrics else None
            for chk in self.checkpoint_list
        ]
        if all([v is None for v in values]):
            raise KeyError("Metric {} does not exist in experiment".format(name))
        return values
