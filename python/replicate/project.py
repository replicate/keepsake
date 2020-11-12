import os
from typing import Dict, Any, Optional

from . import console
from .config import load_config
from .experiment import ExperimentCollection, Experiment
from .repository import repository_for_url, Repository
from .exceptions import ConfigNotFoundError


MAX_SEARCH_DEPTH = 100
DEPRECATED_REPOSITORY_DIR = ".replicate/storage"


class Project:
    """
    Represents a codebase and set of experiments, analogous to a Git repository.
    """

    def __init__(
        self, repository: Optional[str] = None, directory: Optional[str] = None
    ):
        # Project is initialized on import, so don't do anything slow or anything that will raise an exception
        self._directory = directory
        self._repository: Optional[Repository] = None
        self._repository_url = repository
        self._explicit_repository = repository is not None

    @property
    def directory(self) -> str:
        if self._directory is None:
            if self._explicit_repository:
                # we raise an error here rather than in the
                # constructor, because Projects can be used both
                # for writing during training and for analysis.
                # during analysis you don't need a root directory

                raise ValueError(
                    "If you pass the 'repository' argument to Project(), you also need to pass 'directory'"
                )

            self._directory = get_project_dir()
        return self._directory

    def _get_config(self) -> Dict[str, Any]:
        if self._explicit_repository:
            return {"repository": self._repository_url}

        try:
            return load_config(self.directory)
        except ConfigNotFoundError:
            # backwards-compatibility
            # TODO(bfirsh): remove this at some point
            if os.path.exists(os.path.join(self.directory, DEPRECATED_REPOSITORY_DIR)):
                console.warn(
                    f"""replicate.yaml is now required. Create replicate.yaml with this content:

  repository: "file://{DEPRECATED_REPOSITORY_DIR}"
"""
                )
                return {"repository": "file://" + DEPRECATED_REPOSITORY_DIR}
            raise

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
    cwd = os.getcwd()
    directory = cwd
    for _ in range(MAX_SEARCH_DEPTH):
        if os.path.exists(os.path.join(directory, "replicate.yaml")):
            return directory

        # backwards-compatibility
        if os.path.exists(os.path.join(directory, DEPRECATED_REPOSITORY_DIR)):
            return directory

        if directory == "/":
            raise ConfigNotFoundError(
                "replicate.yaml was not found in {} or any of its subdirectories".format(
                    cwd
                )
            )
        directory = os.path.dirname(directory)
    return os.getcwd()
