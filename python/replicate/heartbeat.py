import sys
import datetime
import json
import time
from multiprocessing import Process

from .repository import repository_for_url, Repository
from .metadata import rfc3339_datetime


DEFAULT_REFRESH_INTERVAL = datetime.timedelta(seconds=10)


class Heartbeat:
    def __init__(
        self,
        experiment_id: str,
        repository_url: str,
        path: str,
        refresh_interval: datetime.timedelta = DEFAULT_REFRESH_INTERVAL,
    ):
        self.experiment_id = experiment_id
        self.repository_url = repository_url
        self.path = path
        self.refresh_interval = refresh_interval
        self.process = self.make_process()

    def start(self):
        self.process.start()

    def ensure_running(self):
        if not self.is_alive():
            self.process = self.make_process()
            self.process.start()

    def kill(self):
        self.process.terminate()

    def is_alive(self):
        return self.process.is_alive()

    def make_process(self) -> Process:
        process = Process(target=self.heartbeat_loop)
        process.daemon = True
        return process

    def heartbeat_loop(self):
        # need to instantitate repository here since the gcs
        # client doesn't like multiprocessing:
        # https://github.com/googleapis/google-cloud-python/issues/3501
        repository = repository_for_url(self.repository_url)
        while True:
            self.refresh(repository)
            time.sleep(self.refresh_interval.total_seconds())

    def refresh(self, repository: Repository):
        obj = json.dumps(
            {
                "experiment_id": self.experiment_id,
                "last_heartbeat": rfc3339_datetime(datetime.datetime.utcnow()),
            }
        )
        try:
            repository.put(self.path, obj)
        except Exception as e:  # pylint: disable=broad-except
            sys.stderr.write("Failed to save heartbeat: {}".format(e))
