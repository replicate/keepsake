try:
    # backport is incompatible with 3.7+, so we must use built-in
    from dataclasses import dataclass, InitVar
except ImportError:
    from ._vendor.dataclasses import dataclass, InitVar
import datetime
import os
import io
import tempfile
import json
import sys
import html
from typing import Optional, Dict, Any, List, BinaryIO, TYPE_CHECKING

if sys.version_info >= (3, 8):
    from typing import TypedDict
else:
    from ._vendor.typing_extensions import TypedDict

from . import console
from .json import CustomJSONEncoder
from .hash import random_hash
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

    _experiment: Optional["Experiment"] = None

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

    def checkout(self, output_directory: str, quiet: bool = False):
        """
        Copy files from this checkpoint to the output directory.
        """
        os.makedirs(output_directory, exist_ok=True)

        assert self._experiment is not None
        storage = self._experiment._project._get_storage()
        storage.get_path_tar(self._storage_tar_path(), output_directory)
        storage.get_path_tar(self._experiment._storage_tar_path(), output_directory)
        if not quiet:
            console.info(
                "Copied the files from checkpoint {} to {}".format(
                    self.short_id(), output_directory
                )
            )

    def load(self, path: str) -> BinaryIO:
        """
        Read a single file from this checkpoint into memory.
        Returns a file-like object.
        """
        with tempfile.TemporaryDirectory() as tempdir:
            self.checkout(tempdir, quiet=True)
            with open(os.path.join(tempdir, path), "rb") as f:
                # TODO(andreas): don't load entire file into memory at once, in case it's huge
                out_f = io.BytesIO(f.read())
            return out_f

    def _storage_tar_path(self) -> str:
        return "checkpoints/{}.tar.gz".format(self.id)

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
