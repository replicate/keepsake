import os
from typing import List, Dict, Any

from ._vendor import yaml

from . import console
from .exceptions import ConfigNotFoundError

# TODO (bfirsh): send users to replicate.yaml reference if this is raised!
class ConfigValidationError(Exception):
    pass


def load_config(project_dir: str) -> Dict[str, Any]:
    """
    Loads config from directory
    """
    #  TODO (bfirsh): support replicate.yml too
    try:
        with open(os.path.join(project_dir, "replicate.yaml")) as fh:
            data = yaml.safe_load(fh)
    except FileNotFoundError:
        raise ConfigNotFoundError(
            "replicate.yaml was not found in {}".format(project_dir)
        )
    # Empty file
    if data is None:
        data = {}

    # if replicate is running inside docker and repository is disk,
    # REPLICATE_REPOSITORY is mounted to the value of repository: in
    # replicate.yaml
    if "REPLICATE_REPOSITORY" in os.environ:
        data["repository"] = os.environ["REPLICATE_REPOSITORY"]

    return validate_and_set_defaults(data, project_dir)


# TODO(andreas): more rigorous validation
VALID_KEYS = [
    "repository",
    "python",
    "cuda",
    "python_requirements",
    "install",
    "metrics",
    "storage",
]
REQUIRED_KEYS: List[str] = ["repository"]


def validate_and_set_defaults(data: Dict[str, Any], project_dir: str) -> Dict[str, Any]:
    # TODO (bfirsh): just really simple for now. JSON schema is probably right way (aanand says that is only decent solution)
    if data.get("storage"):
        if data.get("repository"):
            raise ConfigValidationError(
                "Both 'storage' (deprecated) and 'repository' are defined in replicate.yaml, please only use 'repository'"
            )

        console.warn(
            "'storage' is deprecated in replicate.yaml, please use 'repository'"
        )
        data["repository"] = data["storage"]
        del data["storage"]

    defaults = get_default_config()

    for key, value in defaults.items():
        if key not in data:
            data[key] = value

    for key, value in data.items():
        if key not in VALID_KEYS:
            raise ConfigValidationError(
                "The option '{}' is in replicate.yaml, but it is not supported.".format(
                    key
                )
            )

        if key == "repository":
            if not isinstance(value, str):
                raise ConfigValidationError(
                    "The option 'repository' in replicate.yaml needs to be a string."
                )

    # check for required keys last since repository is set from
    # storage for backwards compatibility
    for key in REQUIRED_KEYS:
        if key not in data:
            raise ConfigValidationError(
                "The option '{}' is required in replicate.yaml, but you have no set it.".format(
                    key
                )
            )

    return data


def get_default_config() -> Dict[str, Any]:
    return {}
