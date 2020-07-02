import hashlib
import json
import random

from .hash import random_hash


class Commit(object):
    """
    A snapshot of a training job -- the working directory plus any metadata.
    """

    def __init__(self, experiment, project_dir, metrics):
        self.experiment = experiment
        self.project_dir = project_dir
        self.metrics = metrics

        # TODO (bfirsh): content addressable id
        self.id = random_hash()

    def save(self, storage):
        storage.put_directory(self.get_path(), self.project_dir)
        storage.put(
            self.get_path() + "replicate-metadata.json",
            json.dumps(
                {
                    "id": self.id,
                    "experiment": self.experiment.get_metadata(),
                    "metrics": self.metrics,
                },
                indent=2,
            ),
        )

    def get_path(self):
        return "commits/{}/".format(self.id)
