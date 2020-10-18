from .project import Project, init

default_project = Project()
experiments = default_project.experiments

__all__ = ["default_project", "experiments", "init", "Project"]
