import datetime
import json
import sys
from typing import Dict, Any, Optional, List

from .commit import Commit
from .config import load_config
from .hash import random_hash
from .metadata import rfc3339_datetime
from .project import get_project_dir
from .storage import storage_for_url, Storage
from .heartbeat import Heartbeat


class Experiment:
    def __init__(
        self,
        storage_url: str,
        project_dir: str,
        created: datetime.datetime,
        params: Optional[Dict[str, Any]],
        args: Optional[List[str]],
    ):
        self.storage = storage_for_url(storage_url)
        # TODO: automatically detect workdir (see .project)
        self.project_dir = project_dir
        self.params = params
        self.id = random_hash()
        self.created = created
        self.heartbeat = Heartbeat(
            storage_url, self.get_path() + "replicate-heartbeat.json"
        )

    def save(self):
        self.storage.put(
            self.get_path() + "replicate-metadata.json",
            json.dumps(self.get_metadata(), indent=2),
        )

    def commit(self, metrics: Dict[str, Any]) -> Commit:
        created = datetime.datetime.utcnow()
        commit = Commit(self, self.project_dir, created, metrics)
        commit.save(self.storage)
        self.heartbeat.ensure_running()
        return commit

    def get_metadata(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "created": rfc3339_datetime(self.created),
            "params": self.params,
        }

    def get_path(self):
        return "experiments/{}/".format(self.id)


def init(
    params: Optional[Dict[str, Any]] = None, include_argv: bool = True
) -> Experiment:
    project_dir = get_project_dir()
    config = load_config(project_dir)
    storage_url = config["storage"]
    created = datetime.datetime.utcnow()
    args: Optional[List[str]]
    if include_argv:
        args = sys.argv
    else:
        args = None
    experiment = Experiment(storage_url, project_dir, created, params, args)
    experiment.save()
    experiment.heartbeat.start()
    return experiment
