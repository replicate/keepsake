import time
import json
import os
import datetime
import dateutil
from dateutil.tz import tzutc
import pytest

from replicate.heartbeat import Heartbeat


def test_heartbeat_running(tmpdir):
    tmpdir = tmpdir.dirname
    path = "foo/heartbeat.json"
    heartbeat = Heartbeat(tmpdir, path, datetime.timedelta(seconds=1))
    assert not heartbeat.is_alive()

    heartbeat.start()
    assert heartbeat.is_alive()

    heartbeat.kill()
    time.sleep(0.1)
    assert not heartbeat.is_alive()
    heartbeat.ensure_running()
    assert heartbeat.is_alive()


def test_heartbeat_write(tmpdir):
    tmpdir = tmpdir.dirname
    t1 = datetime.datetime.utcnow().replace(tzinfo=tzutc())

    path = "foo/heartbeat.json"
    heartbeat = Heartbeat(tmpdir, path, datetime.timedelta(seconds=1))
    heartbeat.start()
    time.sleep(2.0)

    with open(os.path.join(tmpdir, "foo", "heartbeat.json")) as f:
        obj = json.loads(f.read())
    last_heartbeat = dateutil.parser.parse(obj["last_heartbeat"])

    t2 = datetime.datetime.utcnow().replace(tzinfo=tzutc())

    assert t1 < last_heartbeat < t2

    time.sleep(2)

    with open(os.path.join(tmpdir, "foo", "heartbeat.json")) as f:
        obj = json.loads(f.read())
    new_last_heartbeat = dateutil.parser.parse(obj["last_heartbeat"])

    assert t1 < last_heartbeat < t2 < new_last_heartbeat
