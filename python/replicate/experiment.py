try:
    # backport is incompatible with 3.7+, so we must use built-in
    from dataclasses import dataclass, InitVar
except ImportError:
    from ._vendor.dataclasses import dataclass, InitVar
import getpass
import os
import datetime
import inspect
import json
import shlex
import sys
from typing import Dict, Any, Optional, Tuple, List
import warnings

from . import console
from .checkpoint import Checkpoint, CheckpointCollection, PrimaryMetric
from .hash import random_hash
from .heartbeat import Heartbeat
from .metadata import rfc3339_datetime


@dataclass
class Experiment:
    """
    A run of a training script.
    """

    project: InitVar[Any]

    id: str
    created: datetime.datetime
    user: str
    host: str
    command: str
    config: dict
    path: Optional[str]
    params: Optional[Dict[str, Any]] = None

    def __post_init__(self, project):
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

        return errors

    @property
    def checkpoints(self) -> CheckpointCollection:
        return CheckpointCollection(self)

    def checkpoint(
        self,
        path: Optional[str],  # this requires an explicit path=None to not save source
        step: Optional[int] = None,
        metrics: Optional[Dict[str, Any]] = None,
        primary_metric: Optional[Tuple[str, str]] = None,
        quiet: bool = False,
        **kwargs,
    ) -> Optional[Checkpoint]:
        """
        Create a checkpoint within this experiment.

        This saves the metrics at this point, and makes a copy of the file or directory passed to `path`, which could be weights or any other artifact.
        """
        if kwargs:
            # FIXME (bfirsh): remove before launch
            s = """Unexpected keyword arguments to checkpoint(): {} 

Metrics must now be passed as a dictionary with the 'metrics' argument.

For example: experiment.checkpoint(path=".", metrics={{...}})

See the docs for more information: https://beta.replicate.ai/docs/python"""
            raise TypeError(s.format(", ".join(kwargs.keys())))

        if path is not None:
            # TODO: Migrate this to validate
            check_path(path)

        # TODO(bfirsh): display warning if primary_metric changes in an experiment
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

        checkpoint = self.checkpoints.create(
            path=path,
            step=step,
            metrics=metrics,
            primary_metric=primary_metric_dict,
            quiet=quiet,
        )

        if self._heartbeat is not None:
            self._heartbeat.ensure_running()

        return checkpoint

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
        }

    def start_heartbeat(self):
        self._heartbeat = Heartbeat(
            experiment_id=self.id,
            storage_url=self._project.config["storage"],
            path="metadata/heartbeats/{}.json".format(self.id),
        )
        self._heartbeat.start()


@dataclass
class ExperimentCollection:
    """
    An object for managing experiments in a project.
    """

    project: Any  # circular import

    def create(self, path=None, params=None) -> Experiment:
        command = " ".join(map(shlex.quote, sys.argv))
        experiment = Experiment(
            project=self.project,
            id=random_hash(),
            created=datetime.datetime.utcnow(),
            path=path,
            params=params,
            config=self.project.config,
            user=os.getenv("REPLICATE_INTERNAL_USER", getpass.getuser()),
            host=os.getenv("REPLICATE_INTERNAL_HOST", ""),
            command=os.getenv("REPLICATE_INTERNAL_COMMAND", command),
        )

        console.info(
            "Creating experiment {}: copying '{}' to '{}'...".format(
                experiment.short_id(), experiment.path, self.project.storage.root_url()
            )
        )

        errors = experiment.validate()
        if errors:
            for error in errors:
                console.error("Not saving experiment: " + error)
            return experiment

        self.project.storage.put(
            "metadata/experiments/{}.json".format(experiment.id),
            json.dumps(experiment.to_json(), indent=2),
        )
        # This is intentionally after uploading the metadata file.
        # When you upload an object to a GCS bucket that doesn't exist, the upload of
        # the first object creates the bucket.
        # If you upload lots of objects in parallel to a bucket that doesn't exist, it
        # causes a race condition, throwing 404s.
        # Hence, uploading the single metadata file is done first.
        # FIXME (bfirsh): this will cause partial experiments if process quits half way through put_path
        if experiment.path is not None:
            source_path = os.path.normpath(
                os.path.join(self.project.dir, experiment.path)
            )
            destination_path = os.path.normpath(
                os.path.join("experiments", experiment.id, experiment.path)
            )
            self.project.storage.put_path(destination_path, source_path)

        return experiment


CHECK_PATH_HELP_TEXT = """

It is relative to the project directory, which is the directory that contains replicate.yaml. You probably just want to set it to path=\".\" to save everything, or path=\"somedir/\" to just save a particular directory.

To learn more, see the documentation: https://beta.replicate.ai/docs/python"""


def check_path(path: str):
    func_name = inspect.stack()[1].function
    # There are few other ways this can break (e.g. "dir/../../") but this will cover most ways users can trip up
    if path.startswith("/") or path.startswith(".."):
        raise ValueError(
            "The path passed to {}() must not start with '..' or '/'.".format(func_name)
            + CHECK_PATH_HELP_TEXT
        )
    if not os.path.exists(path):
        raise ValueError(
            "The path passed to {}() does not exist: {}".format(func_name, path)
            + CHECK_PATH_HELP_TEXT
        )


def init(
    params: Optional[Dict[str, Any]] = None, disable_heartbeat: bool = False, **kwargs,
) -> Experiment:
    """
    Create a new experiment.
    """
    try:
        path = kwargs.pop("path")
    except KeyError:
        warnings.warn(
            "The 'path' argument now needs to be passed to replicate.init() and this will throw an error at some point. "
            "Add 'path=\".\"' to your replicate.init() arguments when you get a chance.",
        )
        path = "."
    if path is not None:
        # TODO: Migrate this to validate
        check_path(path)

    if kwargs:
        # FIXME (bfirsh): remove before launch
        s = """Unexpected keyword arguments to init(): {} 
            
Params must now be passed as a dictionary with the 'params' argument.

For example: replicate.init(path=".", params={{...}})

See the docs for more information: https://beta.replicate.ai/docs/python"""
        raise TypeError(s.format(", ".join(kwargs.keys())))

    # circular import
    from .project import Project

    project = Project()
    experiment = project.experiments.create(path=path, params=params)

    if not disable_heartbeat:
        experiment.start_heartbeat()

    return experiment
