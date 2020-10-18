import os
from typing import Dict, Any, Optional

from .config import load_config
from .experiment import ExperimentCollection, Experiment
from .storage import storage_for_url, Storage


MAX_SEARCH_DEPTH = 100


class Project:
    """
    Represents a codebase and set of experiments, analogous to a Git repository.
    """

    def __init__(self, directory: Optional[str] = None):
        # Project is initialized on import, so don't do anything slow or anything that will raise an exception
        self._directory = directory
        self._config: Optional[Dict[str, Any]] = None
        self._storage: Optional[Storage] = None

    @property
    def directory(self) -> str:
        if self._directory is None:
            self._directory = get_project_dir()
        return self._directory

    def _get_config(self) -> Dict[str, Any]:
        if self._config is None:
            self._config = load_config(self.directory)
        return self._config

    def _get_storage(self) -> Storage:
        if self._storage is None:
            self._storage = storage_for_url(self._get_config()["storage"])
        return self._storage

    @property
    def experiments(self) -> ExperimentCollection:
        return ExperimentCollection(self)


def init(
    path: Optional[str] = None,
    params: Optional[Dict[str, Any]] = None,
    disable_heartbeat: bool = False,
) -> Experiment:
    """
    Create a new experiment.
    """
    project = Project()
    experiment = project.experiments.create(path=path, params=params)

    if not disable_heartbeat:
        experiment.start_heartbeat()

    return experiment


def get_project_dir() -> str:
    """
    Returns the directory of the current project.

    Similar to config.FindConfigPath() in CLI.
    """
    directory = os.getcwd()
    for _ in range(MAX_SEARCH_DEPTH):
        if os.path.exists(os.path.join(directory, "replicate.yaml")):
            return directory
        if directory == "/":
            break
        directory = os.path.dirname(directory)
    return os.getcwd()
