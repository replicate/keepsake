from .project import Project, init
from .version import version as __version__

default_project = Project()
experiments = default_project.experiments

__all__ = ["default_project", "experiments", "init", "Project", "__version__"]
