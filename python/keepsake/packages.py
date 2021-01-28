from typing import Dict
import sys

from . import console


def get_imported_packages() -> Dict[str, str]:
    """
    Returns a list of packages that have been imported, as a {name: version} dict.
    """
    try:
        # Should we vendor pkg_resources? See https://github.com/replicate/keepsake/issues/350
        import pkg_resources
    except ImportError:
        console.warn(
            "Could not import setuptools/pkg_resources, not tracking package versions"
        )
        return {}
    result = {}
    for d in pkg_resources.working_set:
        if is_imported(d.key):
            result[d.key] = d.version
    return result


def is_imported(module_name):
    return module_name in sys.modules
