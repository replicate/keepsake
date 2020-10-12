try:
    import dataclasses
except ImportError:
    from replicate._vendor import dataclasses
import datetime
import json
import os
import pytest  # type: ignore

import replicate
from replicate.exceptions import DoesNotExistError
from replicate.experiment import Experiment
from replicate.project import Project


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
        ".replicate/storage/metadata/experiments/{}.json".format(experiment.id)
    ) as fh:
        metadata = json.load(fh)
    assert len(metadata["checkpoints"]) == 1
    checkpoint_metadata = metadata["checkpoints"][0]
    assert checkpoint_metadata["id"] == checkpoint.id
    assert checkpoint_metadata["step"] == 1
    assert checkpoint_metadata["metrics"] == {"validation_loss": 0.123}
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
        ".replicate/storage/metadata/experiments/{}.json".format(experiment.id)
    ) as fh:
        metadata = json.load(fh)
    assert metadata["checkpoints"][-1]["id"] == checkpoint.id
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


class TestExperiment:
    def test_validate(self):
        kwargs = {
            "project": None,
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

    def test_from_json(self):
        data = {
            "id": "3132f9288bcc09a6b4d283c95a3968379d6b01fcf5d06500e789f90fdb02b7e1",
            "created": "2020-10-07T22:44:06.243914Z",
            "params": {"learning_rate": 0.01, "num_epochs": 100},
            "user": "ben",
            "host": "",
            "command": "train.py",
            "config": {"python": "3.8", "storage": ".replicate/storage/"},
            "path": ".",
            "checkpoints": [],
        }
        exp = Experiment.from_json(None, data)
        assert dataclasses.asdict(exp) == {
            "id": "3132f9288bcc09a6b4d283c95a3968379d6b01fcf5d06500e789f90fdb02b7e1",
            "created": datetime.datetime(2020, 10, 7, 22, 44, 6, 243914),
            "params": {"learning_rate": 0.01, "num_epochs": 100},
            "user": "ben",
            "host": "",
            "command": "train.py",
            "config": {"python": "3.8", "storage": ".replicate/storage/"},
            "path": ".",
            "checkpoints": [],
        }

    def test_checkpoints(self, temp_workdir):
        project = Project()
        experiment = project.experiments.create(path=None, params={"foo": "bar"})
        chk1 = experiment.checkpoint(path=None, metrics={"accuracy": "ok"})
        chk2 = experiment.checkpoint(path=None, metrics={"accuracy": "super"})
        assert len(experiment.checkpoints) == 2
        assert experiment.checkpoints[0].id == chk1.id
        assert experiment.checkpoints[1].id == chk2.id


class TestExperimentCollection:
    def test_get(self, temp_workdir):
        project = Project()
        exp1 = project.experiments.create(path=None, params={"foo": "bar"})
        exp1.checkpoint(path=None, metrics={"accuracy": "wicked"})
        exp2 = project.experiments.create(path=None, params={"foo": "baz"})

        actual_exp = project.experiments.get(exp1.id)
        assert actual_exp.created == exp1.created
        assert len(actual_exp.checkpoints) == 1
        assert actual_exp.checkpoints[0].metrics == {"accuracy": "wicked"}
        # get by prefix
        assert project.experiments.get(exp2.id[:7]).created == exp2.created

        with pytest.raises(DoesNotExistError):
            project.experiments.get("doesnotexist")

    def test_list(self, temp_workdir):
        project = Project()
        exp1 = project.experiments.create(path=None, params={"foo": "bar"})
        exp1.checkpoint(path=None, metrics={"accuracy": "wicked"})
        exp2 = project.experiments.create(path=None, params={"foo": "baz"})

        experiments = project.experiments.list()
        assert len(experiments) == 2
        assert experiments[0].id == exp1.id
        assert len(experiments[0].checkpoints) == 1
        assert experiments[0].checkpoints[0].metrics == {"accuracy": "wicked"}
        assert experiments[1].id == exp2.id
