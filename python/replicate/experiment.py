import hashlib
import json
import random

from .commit import Commit
from .hash import random_hash


class Experiment(object):
    def __init__(self, storage, workdir, params):
        self.storage = storage
        # TODO: automatically detect workdir
        self.workdir = workdir
        self.params = params
        self.id = random_hash()
    
    def save(self):
        self.storage.put(self.get_path() + "replicate-metadata.json", json.dumps(self.get_metadata()))

    def commit(self, metrics):
        commit = Commit(self, self.workdir, metrics)
        commit.save(self.storage)
        return commit

    def get_metadata(self):
        return {
            "id": self.id,
            "params": self.params,
        }

    def get_path(self):
        return "experiments/{}/".format(self.id)



def init(storage, workdir, params=None):
    experiment = Experiment(storage, workdir, params)
    experiment.save()
    return experiment
