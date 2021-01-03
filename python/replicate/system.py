import sys


def get_python_version():
    """
    Returns the Python version of the experiment as a str.
    """
    return ".".join([str(x) for x in sys.version_info[:3]])
