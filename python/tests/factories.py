import datetime
from typing import List, Any, Dict, Optional

from replicate.checkpoint import Checkpoint
from replicate.experiment import Experiment
from replicate.project import Project


def experiment_factory(
    project: Project = None,
    id: str = "e1",
    created: datetime.datetime = datetime.datetime(2020, 1, 1, 1, 1, 1),
    user: str = "bob",
    host: str = "",
    command: str = "",
    config: dict = {},
    replicate_version: str = "0.0.1",
    **kwargs,
):
    return Experiment(
        project=project,
        id=id,
        created=created,
        user=user,
        host=host,
        command=command,
        config=config,
        replicate_version=replicate_version,
        **kwargs,
    )


def checkpoint_factory(
    id: str = "c1",
    created: datetime.datetime = datetime.datetime(2020, 1, 1, 1, 1, 1),
    **kwargs,
):
    return Checkpoint(id=id, created=created, **kwargs)
