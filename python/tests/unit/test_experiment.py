import datetime
import json
import os
import pytest  # type: ignore

import replicate
from replicate.experiment import Experiment

from .common import temp_workdir


def test_init_and_checkpoint(temp_workdir):
    with open("train.py", "w") as fh:
        fh.write("print(1 + 1)")

    with open("README.md", "w") as fh:
        fh.write("Hello")

    # basic experiment
    experiment = replicate.init(
        path=".", params={"learning_rate": 0.002}, disable_heartbeat=True
    )

    assert len(experiment.id) == 64
    with open(
        ".replicate/storage/metadata/experiments/{}.json".format(experiment.id)
    ) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == experiment.id
    assert metadata["params"] == {"learning_rate": 0.002}

    with open(".replicate/storage/experiments/{}/train.py".format(experiment.id)) as fh:
        assert fh.read() == "print(1 + 1)"

    assert os.path.exists(
        ".replicate/storage/experiments/{}/README.md".format(experiment.id)
    )

    # checkpoint with a file
    with open("weights", "w") as fh:
        fh.write("1.2kg")

    checkpoint = experiment.checkpoint(
        path="weights", step=1, metrics={"validation_loss": 0.123}
    )

    assert len(checkpoint.id) == 64
    with open(
        ".replicate/storage/metadata/checkpoints/{}.json".format(checkpoint.id)
    ) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == checkpoint.id
    assert metadata["step"] == 1
    assert metadata["metrics"] == {"validation_loss": 0.123}
    assert metadata["experiment_id"] == experiment.id
    with open(".replicate/storage/checkpoints/{}/weights".format(checkpoint.id)) as fh:
        assert fh.read() == "1.2kg"
    assert (
        os.path.exists(
            ".replicate/storage/checkpoints/{}/train.py".format(checkpoint.id)
        )
        is False
    )

    # checkpoint with a directory
    os.mkdir("data")
    with open("data/weights", "w") as fh:
        fh.write("1.3kg")

    checkpoint = experiment.checkpoint(
        path="data", step=1, metrics={"validation_loss": 0.123}
    )

    with open(
        ".replicate/storage/checkpoints/{}/data/weights".format(checkpoint.id)
    ) as fh:
        assert fh.read() == "1.3kg"
    assert (
        os.path.exists(
            ".replicate/storage/checkpoints/{}/train.py".format(checkpoint.id)
        )
        is False
    )

    # checkpoint with no path
    checkpoint = experiment.checkpoint(
        path=None, step=1, metrics={"validation_loss": 0.123}
    )
    with open(
        ".replicate/storage/metadata/checkpoints/{}.json".format(checkpoint.id)
    ) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == checkpoint.id
    assert (
        os.path.exists(".replicate/storage/checkpoints/{}".format(checkpoint.id))
        is False
    )

    # checkpoint: various path problems
    with pytest.raises(TypeError, match="missing 1 required positional argument"):
        # pylint: disable=no-value-for-parameter
        checkpoint = experiment.checkpoint()
    with pytest.raises(
        ValueError,
        match=r"The path passed to checkpoint\(\) must not start with '..' or '/'.",
    ):
        experiment.checkpoint(path="..")
        experiment.checkpoint(path="/")
    with pytest.raises(
        ValueError, match=r"The path passed to checkpoint\(\) does not exist: blah",
    ):
        experiment.checkpoint(path="blah")

    # experiment with file
    experiment = replicate.init(
        path="train.py", params={"learning_rate": 0.002}, disable_heartbeat=True
    )
    assert os.path.exists(
        ".replicate/storage/experiments/{}/train.py".format(experiment.id)
    )
    assert not os.path.exists(
        ".replicate/storage/experiments/{}/README.md".format(experiment.id)
    )

    # experiment with no path!
    experiment = replicate.init(
        path=None, params={"learning_rate": 0.002}, disable_heartbeat=True
    )
    with open(
        ".replicate/storage/metadata/experiments/{}.json".format(experiment.id)
    ) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == experiment.id
    assert metadata["params"] == {"learning_rate": 0.002}
    assert not os.path.exists(".replicate/storage/experiments/{}".format(experiment.id))

    # experiment: various path problems
    with pytest.raises(
        ValueError,
        match=r"The path passed to init\(\) must not start with '..' or '/'.",
    ):
        replicate.init(path="..")
        replicate.init(path="/")
    with pytest.raises(
        ValueError, match=r"The path passed to init\(\) does not exist: blah",
    ):
        replicate.init(path="blah")


def test_heartbeat(temp_workdir):
    experiment = replicate.init(path=".")
    # Don't write heartbeat
    experiment._heartbeat.kill()
    assert experiment._heartbeat.path == "metadata/heartbeats/{}.json".format(
        experiment.id
    )


class Blah:
    pass


def test_validate():
    kwargs = {
        "_project": None,
        "id": "abc123",
        "created": datetime.datetime.utcnow(),
        "user": "ben",
        "host": "",
        "config": {},
        "command": "",
    }

    experiment = Experiment(path=None, params="lol", **kwargs)
    assert experiment.validate() == ["params must be a dictionary"]

    experiment = Experiment(path=None, params={"foo": Blah()}, **kwargs)
    assert "Failed to serialize the param 'foo' to JSON" in experiment.validate()[0]
