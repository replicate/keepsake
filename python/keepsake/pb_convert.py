import datetime
import json
from typing import List, Dict, Any, Optional, MutableMapping

from google.protobuf import timestamp_pb2

from .servicepb import keepsake_pb2 as pb
from .experiment import Experiment
from .checkpoint import Checkpoint, PrimaryMetric, CheckpointList

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


def timestamp_from_pb(t: timestamp_pb2.Timestamp) -> datetime.datetime:
    return datetime.datetime.fromtimestamp(t.seconds + t.nanos / 1e9)


def checkpoints_from_pb(
    experiment: Experiment,
    checkpoints_pb,  # TODO(andreas): should be RepeatedCompositeFieldContainer[pb.Checkpoint], but that throws TypeError
) -> CheckpointList:
    lst = CheckpointList()
    for chk_pb in checkpoints_pb:
        lst.append(checkpoint_from_pb(experiment, chk_pb))
    return lst


def checkpoint_from_pb(experiment: Experiment, chk_pb: pb.Checkpoint) -> Checkpoint:
    chk = Checkpoint(
        id=chk_pb.id,
        created=timestamp_from_pb(chk_pb.created),
        path=noneable(chk_pb.path),
        step=chk_pb.step,
        metrics=value_map_from_pb(chk_pb.metrics),
        primary_metric=primary_metric_from_pb(chk_pb.primaryMetric),
    )
    chk._experiment = experiment
    return chk


def experiments_from_pb(
    project, experiments_pb: List[pb.Experiment]
) -> List[Experiment]:
    result: List[Experiment] = []
    for exp_pb in experiments_pb:
        result.append(experiment_from_pb(project, exp_pb))
    return result


def experiment_from_pb(project, exp_pb: pb.Experiment) -> Experiment:
    exp = Experiment(
        project=project,
        id=exp_pb.id,
        created=timestamp_from_pb(exp_pb.created),
        user=noneable(exp_pb.user),
        host=noneable(exp_pb.host),
        command=noneable(exp_pb.command),
        config=config_from_pb(exp_pb.config),
        path=noneable(exp_pb.path),
        params=value_map_from_pb(exp_pb.params),
        python_packages=noneable(exp_pb.pythonPackages),
        python_version=noneable(exp_pb.pythonVersion),
        keepsake_version=noneable(exp_pb.keepsakeVersion),
    )
    exp.checkpoints = checkpoints_from_pb(exp, exp_pb.checkpoints)
    return exp


def config_from_pb(conf_pb: Optional[pb.Config]) -> Optional[Dict[str, Any]]:
    if not conf_pb:
        return None
    if not conf_pb.repository and not conf_pb.storage:
        return None
    return {"repository": conf_pb.repository, "storage": conf_pb.storage}


def primary_metric_from_pb(pm_pb: pb.PrimaryMetric,) -> Optional[PrimaryMetric]:
    if not pm_pb.name:
        return None
    if pm_pb.goal == pb.PrimaryMetric.Goal.MAXIMIZE:
        goal = "maximize"
    else:
        goal = "minimize"

    return PrimaryMetric(name=pm_pb.name, goal=goal,)


def value_map_from_pb(
    vm_pb: MutableMapping[str, pb.ParamType]
) -> Optional[Dict[str, Any]]:
    if not vm_pb:
        return None
    return {k: value_from_pb(v) for k, v in vm_pb.items()}


def value_from_pb(value_pb: pb.ParamType) -> Any:
    which = value_pb.WhichOneof("value")
    if which == "boolValue":
        return value_pb.boolValue
    if which == "intValue":
        return value_pb.intValue
    if which == "floatValue":
        return value_pb.floatValue
    if which == "stringValue":
        return value_pb.stringValue
    if which == "objectValueJson":
        return json.loads(value_pb.objectValueJson)


def timestamp_to_pb(t: datetime.datetime) -> timestamp_pb2.Timestamp:
    return timestamp_pb2.Timestamp(
        seconds=int(t.timestamp()), nanos=round((t.timestamp() % 1.0) * 1e9)
    )


def experiment_to_pb(exp: Experiment) -> pb.Experiment:
    return pb.Experiment(
        id=exp.id,
        created=timestamp_to_pb(exp.created),
        user=exp.user,
        host=exp.host,
        command=exp.command,
        config=config_to_pb(exp.config),
        path=exp.path,
        params=value_map_to_pb(exp.params),
        pythonPackages=exp.python_packages,
        pythonVersion=exp.python_version,
        keepsakeVersion=exp.keepsake_version,
        checkpoints=checkpoints_to_pb(exp.checkpoints),
    )


def config_to_pb(conf: Optional[Dict[str, Any]]) -> Optional[pb.Config]:
    if conf is None:
        return None
    return pb.Config(repository=conf["repository"], storage=conf["storage"])


def checkpoints_to_pb(
    checkpoints: Optional[List[Checkpoint]],
) -> Optional[List[pb.Checkpoint]]:
    if checkpoints is None:
        return None
    return [checkpoint_to_pb(chk) for chk in checkpoints]


def checkpoint_to_pb(chk: Checkpoint) -> pb.Checkpoint:
    return pb.Checkpoint(
        id=chk.id,
        created=timestamp_to_pb(chk.created),
        path=chk.path,
        step=chk.step,
        metrics=value_map_to_pb(chk.metrics),
        primaryMetric=primary_metric_to_pb(chk.primary_metric),
    )


def value_map_to_pb(m: Optional[Dict[str, Any]]) -> Optional[Dict[str, pb.ParamType]]:
    if m is None:
        return None
    return {k: value_to_pb(v) for k, v in m.items()}


def value_to_pb(v: Any) -> pb.ParamType:
    if has_numpy:
        if isinstance(v, np.integer):
            return pb.ParamType(intValue=int(v))
        elif isinstance(v, np.floating):
            return pb.ParamType(floatValue=float(v))
        elif isinstance(v, np.ndarray):
            return pb.ParamType(objectValueJson=json.dumps(v.tolist()))
    if _is_torch_tensor(v):
        return pb.ParamType(objectValueJson=json.dumps(v.detach().tolist()))
    if _is_tensorflow_tensor(v):
        return pb.ParamType(objectValueJson=json.dumps(v.numpy().tolist()))
    if isinstance(v, bool):
        return pb.ParamType(boolValue=v)
    if isinstance(v, int):
        return pb.ParamType(intValue=v)
    if isinstance(v, float):
        return pb.ParamType(floatValue=v)
    if isinstance(v, str):
        return pb.ParamType(stringValue=v)
    if isinstance(v, list):
        return pb.ParamType(objectValueJson=json.dumps(v))
    if isinstance(v, dict):
        return pb.ParamType(objectValueJson=json.dumps(v))
    if v is None:
        return pb.ParamType(objectValueJson=json.dumps(v))
    else:
        raise ValueError("Invalid value: %s", v)


def primary_metric_to_pb(pm: Optional[PrimaryMetric]) -> Optional[pb.PrimaryMetric]:
    if pm is None:
        return None
    if pm["goal"] == "maximize":
        goal = pb.PrimaryMetric.Goal.MAXIMIZE
    else:
        goal = pb.PrimaryMetric.Goal.MINIMIZE
    return pb.PrimaryMetric(name=pm["name"], goal=goal)


def noneable(x: Any) -> Optional[Any]:
    if not x:
        return None
    return x
