import hashlib
import json
import random

from .hash import random_hash

class Commit(object):
    """
    A snapshot of a training job -- the working directory plus any metadata.
    """
    def __init__(self, experiment, workdir, metrics):
        self.experiment = experiment
        self.workdir = workdir
        self.metrics = metrics

        # TODO (bfirsh): content addressable id
        self.id = random_hash()
    
    def save(self, storage):
        storage.put_directory(self.get_path(), self.workdir)
        storage.put(self.get_path() + "replicate-metadata.json", json.dumps({
            "id": self.id,
            "experiment": self.experiment.get_metadata(),
            "metrics": self.metrics,
        }))

    def get_path(self):
        return "commits/{}/".format(self.id)
