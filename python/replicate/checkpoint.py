try:
    # backport is incompatible with 3.7+, so we must use built-in
    from dataclasses import dataclass
except ImportError:
    from ._vendor.dataclasses import dataclass
import datetime
import os
import json
import sys
from typing import Optional, Dict, Any, List

if sys.version_info >= (3, 8):
    from typing import TypedDict
else:
    from ._vendor.typing_extensions import TypedDict

from . import console
from .hash import random_hash
from .metadata import rfc3339_datetime


# We load numpy but not torch or tensorflow because numpy loads very fast and
# they're probably using it anyway
# fmt: off
try:
    import numpy as np  # type: ignore
    has_numpy = True
except ImportError:
    has_numpy = False
# fmt: on

# Tensorflow takes a solid 10 seconds to import on a modern Macbook Pro, so instead of importing,
# do this instead
def _is_tensorflow_tensor(obj):
    # e.g. __module__='tensorflow.python.framework.ops', __name__='EagerTensor'
    return (
        obj.__class__.__module__.split(".")[0] == "tensorflow"
        and "Tensor" in obj.__class__.__name__
    )


def _is_torch_tensor(obj):
    return (obj.__class__.__module__, obj.__class__.__name__) == ("torch", "Tensor")


class CustomJSONEncoder(json.JSONEncoder):
    def default(self, o):
        if has_numpy:
            if isinstance(o, np.integer):
                return int(o)
            elif isinstance(o, np.floating):
                return float(o)
            elif isinstance(o, np.ndarray):
                return o.tolist()
        if _is_torch_tensor(o):
            return o.detach().tolist()
        if _is_tensorflow_tensor(o):
            return o.numpy().tolist()
        return json.JSONEncoder.default(self, o)


class PrimaryMetric(TypedDict):
    name: str
    goal: str


@dataclass
class Checkpoint(object):
    """
    A checkpoint within an experiment. It represents the metrics and the file or directory specified by `path` at a point in time during the experiment.
    """

    experiment: Any  # circular import

    id: str
    created: datetime.datetime
    path: Optional[str] = None
    step: Optional[int] = None
    metrics: Optional[Dict[str, Any]] = None
    primary_metric: Optional[PrimaryMetric] = None

    def short_id(self) -> str:
        return self.id[:7]

    def to_json(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "created": rfc3339_datetime(self.created),
            "experiment_id": self.experiment.id,
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

        return errors


@dataclass
class CheckpointCollection:
    """
    An object for managing checkpoints within an experiment.
    """

    experiment: Any  # circular import

    def create(
        self,
        path: Optional[str] = None,
        step: Optional[int] = None,
        metrics: Optional[Dict[str, Any]] = None,
        primary_metric: Optional[PrimaryMetric] = None,
        quiet: bool = False,
    ) -> Checkpoint:
        project = self.experiment._project

        checkpoint = Checkpoint(
            experiment=self.experiment,
            id=random_hash(),
            created=datetime.datetime.utcnow(),
            path=path,
            step=step,
            metrics=metrics,
            primary_metric=primary_metric,
        )
        if not quiet:
            console.info(
                "Creating checkpoint {}: copying '{}' to '{}'...".format(
                    checkpoint.short_id(), checkpoint.path, project.storage.root_url(),
                )
            )

        errors = checkpoint.validate()
        if errors:
            for error in errors:
                console.error("Not saving checkpoint: " + error)
            return checkpoint

        project.storage.put(
            "metadata/checkpoints/{}.json".format(checkpoint.id),
            json.dumps(checkpoint.to_json(), indent=2, cls=CustomJSONEncoder),
        )
        # FIXME (bfirsh): this will cause partial checkpoints if process quits half way through put_path
        if checkpoint.path is not None:
            source_path = os.path.normpath(os.path.join(project.dir, checkpoint.path))
            destination_path = os.path.normpath(
                os.path.join("checkpoints", checkpoint.id, checkpoint.path)
            )
            project.storage.put_path(destination_path, source_path)

        return checkpoint
