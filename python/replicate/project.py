import os
from typing import Dict, Any, Optional

from .config import load_config
from .experiment import ExperimentCollection, Experiment
from .repository import repository_for_url, Repository


MAX_SEARCH_DEPTH = 100


class Project:
    """
    Represents a codebase and set of experiments, analogous to a Git repository.
    """

    def __init__(self, directory: Optional[str] = None):
        # Project is initialized on import, so don't do anything slow or anything that will raise an exception
        self._directory = directory
        self._config: Optional[Dict[str, Any]] = None
        self._repository: Optional[Repository] = None
        self._repository_url: Optional[str] = None

    @property
    def directory(self) -> str:
        if self._directory is None:
            self._directory = get_project_dir()
        return self._directory

    def _get_config(self) -> Dict[str, Any]:
        return load_config(self.directory)

    def _get_repository(self) -> Repository:
        reload_repository = self._repository is None
        if self._repository_url is not None:
            config = self._get_config()
            if config["repository"] != self._repository_url:
                reload_repository = True
                self._repository_url = config["repository"]

        if reload_repository:
            if self._repository_url is None:
                config = self._get_config()
                self._repository_url = config["repository"]

            self._repository = repository_for_url(self._repository_url)

        return self._repository  # type: ignore

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
    return project.experiments.create(
        path=path, params=params, disable_heartbeat=disable_heartbeat
    )


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
