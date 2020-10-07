import os

from .config import load_config
from .experiment import ExperimentCollection
from .storage import storage_for_url


MAX_SEARCH_DEPTH = 100


class Project:
    """
    Represents a codebase and set of experiments, analogous to a Git repository.
    """

    def __init__(self):
        self.dir = get_project_dir()
        self.config = load_config(self.dir)
        self.storage = storage_for_url(self.config["storage"])

    @property
    def experiments(self) -> ExperimentCollection:
        return ExperimentCollection(self)


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
