import hashlib
import json
import random

from .commit import Commit
from .config import load_config
from .hash import random_hash
from .project import get_project_dir
from .storage import DiskStorage


class Experiment(object):
    def __init__(self, storage, project_dir, params):
        self.storage = storage
        # TODO: automatically detect workdir
        self.project_dir = project_dir
        self.params = params
        self.id = random_hash()

    def save(self):
        self.storage.put(
            self.get_path() + "replicate-metadata.json",
            json.dumps(self.get_metadata(), indent=2),
        )

    def commit(self, metrics):
        commit = Commit(self, self.project_dir, metrics)
        commit.save(self.storage)
        return commit

    def get_metadata(self):
        return {
            "id": self.id,
            "params": self.params,
        }

    def get_path(self):
        return "experiments/{}/".format(self.id)


def init(params=None):
    project_dir = get_project_dir()
    config = load_config(project_dir)
    storage = DiskStorage(config["storage"])
    experiment = Experiment(storage, project_dir, params)
    experiment.save()
    return experiment
