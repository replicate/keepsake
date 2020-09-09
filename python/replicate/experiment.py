import getpass
import os
import datetime
import json
import html
import sys
from typing import Dict, Any, Optional, List, Tuple

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
        params: Optional[Dict[str, Any]],
        disable_heartbeat: bool = False,
    ):
        self.config = config
        storage_url = config["storage"]
        self.storage = storage_for_url(storage_url)
        # TODO: automatically detect workdir (see .project)
        self.project_dir = project_dir
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
        self.storage.put_directory("experiments/{}/".format(self.id), self.project_dir)

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
            raise TypeError(
                """Metrics must now be passed as a dictionary with the 'metrics' argument.

    For example: experiment.checkpoint(path=".", metrics={...})

    See the docs for more information: https://beta.replicate.ai/docs/python"""
            )
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
    disable_heartbeat: bool = False, params: Optional[Dict[str, Any]] = None, **kwargs
) -> Experiment:
    if kwargs:
        # FIXME (bfirsh): remove before launch
        raise TypeError(
            """Params must now be passed as a dictionary with the 'params' argument.

For example: replicate.init(params={...})

See the docs for more information: https://beta.replicate.ai/docs/python"""
        )
    project_dir = get_project_dir()
    config = load_config(project_dir)
    created = datetime.datetime.utcnow()
    experiment = Experiment(
        config=config,
        project_dir=project_dir,
        created=created,
        params=params,
        disable_heartbeat=disable_heartbeat,
    )
    experiment.save()
    if not disable_heartbeat:
        experiment.heartbeat.start()
    return experiment


class ExperimentList(list):
    def _repr_html_(self):
        headings = ["id", "created", "user", "host", "params", "command"]
        out = ["<table>"]
        out.append("<tr>")
        for h in headings:
            out.append("<th>")
            out.append(html.escape(h))
            out.append("</th>")
        out.append("</tr>")
        for experiment in self:
            out.append("<tr>")
            for h in headings:
                d = experiment.get(h)
                if isinstance(d, dict):
                    d = str(d)
                if h == "id":
                    d = d[:7]
                out.append("<th>")
                out.append(html.escape(d))
                out.append("</th>")
            out.append("</tr>")
        out.append("</table>")
        return "".join(out)


# TODO: maybe define this in a class, then set to replicate.list in __init__.py so we're not overriding a built-in
def list():
    project_dir = get_project_dir()
    config = load_config(project_dir)
    storage_url = config["storage"]
    storage = storage_for_url(storage_url)
    result = ExperimentList([])
    for info in storage.list("metadata/experiments/"):
        # FIXME: list should return full path to match Go API
        result.append(json.loads(storage.get("metadata/experiments/" + info["name"])))
    return result


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
