import os
from pathlib import Path

ROOT_DIRECTORY = Path(__file__).parent.parent.parent


def get_env():
    """
    Returns environment for running Replicate commands in
    """
    env = os.environ
    env["PATH"] = "/usr/local/bin:" + os.environ["PATH"]
    return env
