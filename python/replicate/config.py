import os
import yaml
from typing import List

# TODO (bfirsh): send users to replicate.yaml reference if this is raised!
class ConfigValidationError(Exception):
    pass


def load_config(project_dir):
    """
    Loads config from directory
    """
    #  TODO (bfirsh): support replicate.yml too
    try:
        with open(os.path.join(project_dir, "replicate.yaml")) as fh:
            data = yaml.safe_load(fh)
    except FileNotFoundError:
        data = {}
    # Empty file
    if data is None:
        data = {}

    # if replicate is running inside docker and storage is disk,
    # REPLICATE_STORAGE is mounted to the value of storage: in
    # replicate.yaml
    if "REPLICATE_STORAGE" in os.environ:
        data["storage"] = os.environ["REPLICATE_STORAGE"]

    return validate_and_set_defaults(data)


# TODO(andreas): more rigorous validation
VALID_KEYS = ["storage", "python", "cuda", "python_requirements", "install"]
REQUIRED_KEYS: List[str] = []
DEFAULTS = {"storage": ".replicate/storage/", "python": "3.7"}


def validate_and_set_defaults(data):
    # TODO (bfirsh): just really simple for now. JSON schema is probably right way (aanand says that is only decent solution)
    for key in REQUIRED_KEYS:
        if key not in data:
            raise ConfigValidationError(
                "The option '{}' is required in replicate.yaml, but you have no set it.".format(
                    key
                )
            )

    for key, value in DEFAULTS.items():
        if key not in data:
            data[key] = value

    for key, value in data.items():
        if key not in VALID_KEYS:
            raise ConfigValidationError(
                "The option '{}' is in replicate.yaml, but it is not supported.".format(
                    key
                )
            )

        if key == "storage":
            if not isinstance(value, str):
                raise ConfigValidationError(
                    "The option 'storage' in replicate.yaml needs to be a string."
                )

    return data
