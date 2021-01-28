import datetime
from typing import List, Any, Dict, Optional

from keepsake.checkpoint import Checkpoint, CheckpointList
from keepsake.experiment import Experiment
from keepsake.project import Project


def experiment_factory(
    project: Project = None,
    id: str = "e1",
    created: datetime.datetime = datetime.datetime(2020, 1, 1, 1, 1, 1),
    user: str = "bob",
    host: str = "",
    command: str = "",
    config: dict = {},
    keepsake_version: str = "0.0.1",
    checkpoints: Optional[List[Checkpoint]] = None,
    **kwargs,
):
    if checkpoints is not None:
        checkpoints = CheckpointList(checkpoints)

    return Experiment(
        project=project,
        id=id,
        created=created,
        user=user,
        host=host,
        command=command,
        config=config,
        keepsake_version=keepsake_version,
        checkpoints=checkpoints,
        **kwargs,
    )


def checkpoint_factory(
    id: str = "c1",
    created: datetime.datetime = datetime.datetime(2020, 1, 1, 1, 1, 1),
    **kwargs,
):
    return Checkpoint(id=id, created=created, **kwargs)
