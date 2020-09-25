import os
from pathlib import Path

ROOT_DIRECTORY = Path(__file__).parent.parent.parent


def get_env():
    """
    Returns environment for running Replicate commands in
    """
    env = os.environ
    env["PATH"] = "/usr/local/bin:" + os.environ["PATH"]

    dist = ROOT_DIRECTORY / "python/dist"
    globs = dist.glob("replicate-*-py3-none-manylinux1_x86_64.whl")
    env["REPLICATE_DEV_PYTHON_PACKAGE"] = str(list(globs)[0])
    return env
