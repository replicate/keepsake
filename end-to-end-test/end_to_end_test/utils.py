import os


def get_env():
    """
    Returns environment for running Replicate commands in
    """
    env = os.environ
    env["PATH"] = "/usr/local/bin:" + os.environ["PATH"]
    env["REPLICATE_DEV_PYTHON_SOURCE"] = os.path.join(
        os.path.dirname(os.path.realpath(__file__)), "../python"
    )
    return env
