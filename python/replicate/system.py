import sys


def get_python_version():
    """
    Returns the Python version of the experiment as a str.
    """
    return sys.version.split(" ")[0]
