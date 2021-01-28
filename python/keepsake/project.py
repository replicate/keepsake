try:
    # backport is incompatible with 3.7+, so we must use built-in
    from dataclasses import dataclass
except ImportError:
    from ._vendor.dataclasses import dataclass
import os
from typing import Dict, Any, Optional
import json

from . import console
from .daemon import Daemon
from .experiment import ExperimentCollection, Experiment


MAX_SEARCH_DEPTH = 100


@dataclass
class ProjectSpec:
    """
    Project-level storage configuration.
    """

    version: int

    @classmethod
    def from_json(cls, data: Dict[str, Any]) -> "ProjectSpec":
        return ProjectSpec(**data)

    def to_json(self):
        return json.dumps({"version": self.version}, indent=2)


class Project:
    """
    Represents a codebase and set of experiments, analogous to a Git repository.
    """

    def __init__(
        self,
        repository: Optional[str] = None,
        directory: Optional[str] = None,
        debug: bool = False,
    ):
        # Project is initialized on import, so don't do anything slow or anything that will raise an exception
        self.directory = directory
        self.repository = repository
        self._daemon_instance: Optional[Daemon] = None
        self._debug = debug

    @property
    def experiments(self) -> ExperimentCollection:
        return ExperimentCollection(self)

    def _daemon(self) -> Daemon:
        if self._daemon_instance is None:
            self._daemon_instance = Daemon(self, debug=self._debug)
        return self._daemon_instance


def init(
    path: Optional[str] = None,
    params: Optional[Dict[str, Any]] = None,
    disable_heartbeat: bool = False,
    debug: bool = False,
) -> Experiment:
    """
    Create a new experiment.
    """
    project = Project(debug=debug)
    return project.experiments.create(
        path=path, params=params, disable_heartbeat=disable_heartbeat
    )
