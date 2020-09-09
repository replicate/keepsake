import os

MAX_SEARCH_DEPTH = 100


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
