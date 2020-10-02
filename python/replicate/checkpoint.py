import datetime
import os
import json
from typing import Optional, Dict, Any, List

from . import console
from .hash import random_hash
from .metadata import rfc3339_datetime
from .storage import Storage


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


class Checkpoint(object):
    """
    A snapshot of a training job -- the working directory plus any metadata.
    """

    def __init__(
        self,
        experiment,  # can't type annotate due to circular import
        path: Optional[str] = None,
        step: Optional[int] = None,
        metrics: Optional[Dict[str, Any]] = None,
        primary_metric_name: Optional[str] = None,
        primary_metric_goal: Optional[str] = None,
    ):
        self.experiment = experiment
        self.path = path
        self.step = step
        self.metrics = metrics
        self.primary_metric_name = primary_metric_name
        self.primary_metric_goal = primary_metric_goal

        # TODO (bfirsh): content addressable id
        self.id = random_hash()
        self.created = datetime.datetime.utcnow()

    def short_id(self):
        return self.id[:7]

    def save(self, storage: Storage):
        errors = self.validate()
        if errors:
            for error in errors:
                console.error("Not saving checkpoint: " + error)
            return

        obj = {
            "id": self.id,
            "created": rfc3339_datetime(self.created),
            "experiment_id": self.experiment.id,
            "path": self.path,
            "metrics": self.metrics,
        }
        if (
            self.primary_metric_name is not None
            and self.primary_metric_goal is not None
        ):
            obj["primary_metric"] = {
                "name": self.primary_metric_name,
                "goal": self.primary_metric_goal,
            }

        if self.step is not None:
            obj["step"] = self.step
        storage.put(
            "metadata/checkpoints/{}.json".format(self.id),
            json.dumps(obj, indent=2, cls=CustomJSONEncoder),
        )
        # FIXME (bfirsh): this will cause partial checkpoints if process quits half way through put_path
        if self.path is not None:
            source_path = os.path.normpath(
                os.path.join(self.experiment.project_dir, self.path)
            )
            destination_path = os.path.normpath(
                os.path.join("checkpoints", self.id, self.path)
            )
            storage.put_path(destination_path, source_path)

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
            and self.primary_metric_name is not None
            and self.primary_metric_name not in self.metrics
        ):
            errors.append(
                "Primary metric '{}' is not defined in metrics".format(
                    self.primary_metric_name
                )
            )

        if self.primary_metric_goal is not None and self.primary_metric_goal.lower() not in (
            "maximize",
            "minimize",
        ):
            errors.append(
                "Primary metric goal must be either 'maximize' or 'minimize'. "
                "For example: ('loss', 'minimize')"
            )

        return errors
