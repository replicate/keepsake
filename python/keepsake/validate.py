import os

from . import constants

CHECK_PATH_HELP_TEXT = """

It is relative to the project directory, which is the directory that contains keepsake.yaml. You probably just want to set it to path=\".\" to save everything, or path=\"somedir/\" to just save a particular directory.

To learn more, see the documentation: {}""".format(
    constants.PYTHON_REFERENCE_DOCS_URL
)


def check_path(thing: str, path: str):
    errors = []
    # There are few other ways this can break (e.g. "dir/../../") but this will cover most ways users can trip up
    if path.startswith("/") or path.startswith(".."):
        errors.append(
            f"The path passed to the {thing} must not start with '..' or '/'."
            + CHECK_PATH_HELP_TEXT
        )
    if not os.path.exists(path):
        errors.append(
            f"The path passed to the {thing} does not exist: {path}"
            + CHECK_PATH_HELP_TEXT
        )
    return errors
