try:
    # backport is incompatible with 3.7+, so we must use built-in
    from dataclasses import dataclass, InitVar, field
except ImportError:
    from ._vendor.dataclasses import dataclass, InitVar, field
import getpass
import os
import datetime
import json
import shlex
import sys
from typing import Dict, Any, Optional, Tuple, List, TYPE_CHECKING

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
    def from_json(self, project: "Project", data: Dict[str, Any]) -> "Experiment":
        data = data.copy()
        data["created"] = parse_rfc3339(data["created"])
        data["checkpoints"] = [
            Checkpoint.from_json(d) for d in data.get("checkpoints", [])
        ]
        return Experiment(project=project, **data)

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
            "checkpoints": [c.to_json() for c in self.checkpoints],
        }

    def start_heartbeat(self):
        self._heartbeat = Heartbeat(
            experiment_id=self.id,
            storage_url=self._project._get_config()["storage"],
            path=self._heartbeat_path(),
        )
        self._heartbeat.start()

    def delete(self):
        # TODO(andreas): this logic should probably live in go,
        # which could then also be parallelized easily
        storage = self._project._get_storage()
        for checkpoint in self.checkpoints:
            console.info("Deleting checkpoint: {}".format(checkpoint.id))
            storage.delete(checkpoint._storage_tar_path())
        console.info("Deleting experiment: {}".format(self.id))
        storage.delete(self._heartbeat_path())
        storage.delete(self._storage_tar_path())
        storage.delete(self._metadata_path())

    def _heartbeat_path(self) -> str:
        return "metadata/heartbeats/{}.json".format(self.id)

    def _storage_tar_path(self) -> str:
        return "experiments/{}.tar.gz".format(self.id)

    def _metadata_path(self) -> str:
        return "metadata/experiments/{}.json".format(self.id)


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

    def list(self) -> List[Experiment]:
        """
        Return all experiments for a project, sorted by creation date.
        """
        storage = self.project._get_storage()
        result: List[Experiment] = []
        for path in storage.list("metadata/experiments/"):
            data = json.loads(storage.get(path))
            result.append(Experiment.from_json(self.project, data))
        result.sort(key=lambda e: e.created)
        return result
