import getpass
import os
import datetime
import json
import sys
from typing import Dict, Any, Optional, List, Tuple, overload
import warnings

from .checkpoint import Checkpoint
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
        path: str,
        params: Optional[Dict[str, Any]],
        disable_heartbeat: bool = False,
    ):
        self.config = config
        storage_url = config["storage"]
        self.storage = storage_for_url(storage_url)
        # TODO: automatically detect workdir (see .project)
        self.project_dir = project_dir
        self.path = path
        self.params = params
        self.id = random_hash()
        self.created = created
        self.disable_heartbeat = disable_heartbeat
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
        # This is intentionally after uploading the metadata file.
        # When you upload an object to a GCS bucket that doesn't exist, the upload of
        # the first object creates the bucket.
        # If you upload lots of objects in parallel to a bucket that doesn't exist, it
        # causes a race condition, throwing 404s.
        # Hence, uploading the single metadata file is done first.
        # FIXME (bfirsh): this will cause partial experiments if process quits half way through put_directory
        if self.path is not None:
            source_path = os.path.normpath(os.path.join(self.project_dir, self.path))
            destination_path = os.path.normpath(
                os.path.join("experiments", self.id, self.path)
            )
            if os.path.isfile(source_path):
                with open(os.path.join(source_path), "rb") as fh:
                    data = fh.read()
                self.storage.put(destination_path, data)
            else:
                self.storage.put_directory(destination_path, source_path)

    def checkpoint(
        self,
        path: Optional[str],  # this requires an explicit path=None to not save source
        step: Optional[int] = None,
        metrics: Optional[Dict[str, Any]] = None,
        primary_metric: Optional[Tuple[str, str]] = None,
        **kwargs,
    ) -> Checkpoint:
        if kwargs:
            # FIXME (bfirsh): remove before launch
            s = """Unexpected keyword arguments to init(): {} 

Metrics must now be passed as a dictionary with the 'metrics' argument.

For example: experiment.checkpoint(path=".", metrics={{...}})

See the docs for more information: https://beta.replicate.ai/docs/python"""
            raise TypeError(s.format(", ".join(kwargs.keys())))

        created = datetime.datetime.utcnow()
        # TODO(bfirsh): display warning if primary_metric changes in an experiment
        primary_metric_name: Optional[str] = None
        primary_metric_goal: Optional[str] = None
        if primary_metric is not None:
            if len(primary_metric) != 2:
                raise ValueError(
                    "primary_metric must be a tuple of (name, goal), where name corresponds to a metric key, and goal is either 'maximize' or 'minimize'"
                )
            primary_metric_name, primary_metric_goal = primary_metric

        checkpoint = Checkpoint(
            experiment=self,
            project_dir=self.project_dir,
            path=path,
            created=created,
            step=step,
            metrics=metrics,
            primary_metric_name=primary_metric_name,
            primary_metric_goal=primary_metric_goal,
        )
        checkpoint.save(self.storage)
        if not self.disable_heartbeat:
            self.heartbeat.ensure_running()
        return checkpoint

    def get_metadata(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "created": rfc3339_datetime(self.created),
            "params": self.params,
            "user": self.get_user(),
            "host": self.get_host(),
            "command": self.get_command(),
            "config": self.config,
            "path": self.path,
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
        return ""

    def get_command(self) -> str:
        return os.environ.get("REPLICATE_COMMAND", " ".join(sys.argv))


def init(
    params: Optional[Dict[str, Any]] = None, disable_heartbeat: bool = False, **kwargs,
) -> Experiment:
    try:
        path = kwargs.pop("path")
    except KeyError:
        warnings.warn(
            "The 'path' argument now needs to be passed to replicate.init() and this will throw an error at some point. "
            "Add 'path=\".\"' to your replicate.init() arguments when you get a chance.",
        )
        path = "."
    # TODO (bfirsh): fail if path is ../

    if kwargs:
        # FIXME (bfirsh): remove before launch
        s = """Unexpected keyword arguments to init(): {} 
            
Params must now be passed as a dictionary with the 'params' argument.

For example: replicate.init(path=".", params={{...}})

See the docs for more information: https://beta.replicate.ai/docs/python"""
        raise TypeError(s.format(", ".join(kwargs.keys())))
    project_dir = get_project_dir()
    config = load_config(project_dir)
    created = datetime.datetime.utcnow()
    experiment = Experiment(
        config=config,
        project_dir=project_dir,
        created=created,
        path=path,
        params=params,
        disable_heartbeat=disable_heartbeat,
    )
    experiment.save()
    if not disable_heartbeat:
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
