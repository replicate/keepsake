import urllib
import urllib.error
import getpass
import os
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
        config: dict,
        project_dir: str,
        created: datetime.datetime,
        params: Optional[Dict[str, Any]],
    ):
        self.config = config
        storage_url = config["storage"]
        self.storage = storage_for_url(storage_url)
        # TODO: automatically detect workdir (see .project)
        self.project_dir = project_dir
        self.params = params
        self.id = random_hash()
        self.created = created
        self.heartbeat = Heartbeat(
            experiment_id=self.id,
            storage_url=storage_url,
            path="metadata/heartbeats/{}.json".format(self.id),
        )

    def save(self):
        self.storage.put(
            "metadata/experiments/{}.json".format(self.id),
            json.dumps(self.get_metadata(), indent=2),
        )

    def commit(
        self, step: Optional[int] = None, options: Optional[Any] = None, **kwargs
    ) -> Commit:
        options = set_option_defaults(options, {})
        created = datetime.datetime.utcnow()
        commit = Commit(
            experiment=self,
            project_dir=self.project_dir,
            created=created,
            step=step,
            labels=kwargs,
        )
        commit.save(self.storage)
        self.heartbeat.ensure_running()
        return commit

    def get_metadata(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "created": rfc3339_datetime(self.created),
            "params": self.params,
            "user": self.get_user(),
            "host": self.get_host(),
            "command": self.get_command(),
            "config": self.config,
        }

    def get_user(self) -> str:
        user = os.environ.get("REPLICATE_USER")
        if user is not None:
            return user
        return getpass.getuser()

    def get_host(self) -> str:
        host = os.environ.get("REPLICATE_HOST")
        if host is not None:
            return host
        try:
            external_ip = (
                # FIXME: check this has a short timeout
                urllib.request.urlopen("https://ident.me")
                .read()
                .decode("utf8")
            )
            return external_ip
        except urllib.error.URLError as e:
            sys.stderr.write("Failed to determine external IP, got error: {}".format(e))
            return ""

    def get_command(self) -> str:
        return os.environ.get("REPLICATE_COMMAND", " ".join(sys.argv))


def init(options: Optional[Dict[str, Any]] = None, **kwargs) -> Experiment:
    options = set_option_defaults(options, {})
    project_dir = get_project_dir()
    config = load_config(project_dir)
    created = datetime.datetime.utcnow()
    experiment = Experiment(
        config=config, project_dir=project_dir, created=created, params=kwargs,
    )
    experiment.save()
    experiment.heartbeat.start()
    return experiment


def set_option_defaults(
    options: Optional[Dict[str, Any]], defaults: Dict[str, Any]
) -> Dict[str, Any]:
    if options is None:
        options = {}
    else:
        options = options.copy()
    for name, value in defaults.items():
        if name not in options:
            options[name] = value
    invalid_options = set(options) - set(defaults)
    if invalid_options:
        raise ValueError(
            "Invalid option{}: {}".format(
                "s" if len(invalid_options) > 1 else "", ", ".join(invalid_options)
            )
        )
    return options
