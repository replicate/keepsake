import hashlib
import json
import random
from typing import Optional, Dict, Any

from .hash import random_hash
from .storage import Storage


class Commit(object):
    """
    A snapshot of a training job -- the working directory plus any metadata.
    """

    def __init__(
        self,
        experiment,  # can't type annotate due to circular import
        project_dir: str,
        timestamp: float,
        metrics: Dict[str, Any],
    ):
        self.experiment = experiment
        self.project_dir = project_dir
        self.timestamp = timestamp
        self.metrics = metrics

        # TODO (bfirsh): content addressable id
        self.id = random_hash()

    def save(self, storage: Storage):
        storage.put_directory(self.get_path(), self.project_dir)
        storage.put(
            self.get_path() + "replicate-metadata.json",
            json.dumps(
                {
                    "id": self.id,
                    "timestamp": self.timestamp,
                    "experiment": self.experiment.get_metadata(),
                    "metrics": self.metrics,
                },
                indent=2,
            ),
        )

    def get_path(self) -> str:
        return "commits/{}/".format(self.id)
