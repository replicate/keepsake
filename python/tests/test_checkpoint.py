try:
    import dataclasses
except (ImportError, ModuleNotFoundError):
    from keepsake._vendor import dataclasses
import datetime
import time
import os
import pytest
from waiting import wait

from keepsake.checkpoint import Checkpoint, CheckpointList
from keepsake.exceptions import DoesNotExist
from keepsake.project import Project

from tests.factories import experiment_factory, checkpoint_factory


class Blah:
    pass


class TestCheckpoint:
    def test_validate(self):
        kwargs = {
            "id": "def456",
            "created": datetime.datetime.utcnow(),
        }

        checkpoint = Checkpoint(path=123, **kwargs)
        assert checkpoint.validate() == ["path must be a string"]

        checkpoint = Checkpoint(step="lol", **kwargs)
        assert checkpoint.validate() == ["step must be an integer"]

        checkpoint = Checkpoint(metrics="lol", **kwargs)
        assert checkpoint.validate() == ["metrics must be a dictionary"]

        checkpoint = Checkpoint(metrics={"foo": Blah()}, **kwargs)
        assert (
            "Failed to serialize the metric 'foo' to JSON" in checkpoint.validate()[0]
        )

        checkpoint = Checkpoint(
            metrics={"foo": "bar"},
            primary_metric={"name": "baz", "goal": "maximize"},
            **kwargs
        )
        assert checkpoint.validate() == [
            "Primary metric 'baz' is not defined in metrics"
        ]

        checkpoint = Checkpoint(
            metrics={"foo": "bar"},
            primary_metric={"name": "foo", "goal": "maximilize"},
            **kwargs
        )
        assert (
            "Primary metric goal must be either 'maximize' or 'minimize'"
            in checkpoint.validate()[0]
        )

        checkpoint = Checkpoint(path="..", **kwargs)
        assert (
            "The path passed to the checkpoint must not start with '..' or '/'."
            in checkpoint.validate()[0]
        )
        checkpoint = Checkpoint(path="/", **kwargs)
        assert (
            "The path passed to the checkpoint must not start with '..' or '/'."
            in checkpoint.validate()[0]
        )
        checkpoint = Checkpoint(path="blah", **kwargs)
        assert (
            "The path passed to the checkpoint does not exist: blah"
            in checkpoint.validate()[0]
        )

    def test_from_json(self, temp_workdir):
        data = {
            "id": "a1a973fcbead997a3b253c960b9fad1bb1386681beefd7ba8703e25548beb11c",
            "created": "2020-10-07T22:44:06.532785Z",
            "path": "model.pth",
            "metrics": {"loss": 0.9042219519615173, "accuracy": 0.8666666746139526},
            "primary_metric": {"name": "loss", "goal": "minimize"},
            "step": 7,
        }
        checkpoint = Checkpoint.from_json(data)
        assert dataclasses.asdict(checkpoint) == {
            "id": "a1a973fcbead997a3b253c960b9fad1bb1386681beefd7ba8703e25548beb11c",
            "created": datetime.datetime(2020, 10, 7, 22, 44, 6, 532785),
            "path": "model.pth",
            "metrics": {"loss": 0.9042219519615173, "accuracy": 0.8666666746139526},
            "primary_metric": {"name": "loss", "goal": "minimize"},
            "step": 7,
        }

    def test_checkout(self, temp_workdir, tmpdir_factory):
        project = Project()
        with open("foo.txt", "w") as f:
            f.write("foo")

        with open("keepsake.yaml", "w") as f:
            f.write("repository: file://.keepsake/")

        exp = project.experiments.create(
            path=".", params={"foo": "bar"}, disable_heartbeat=True
        )
        with open("bar.txt", "w") as f:
            f.write("bar")
        chk = exp.checkpoint(path="bar.txt", metrics={"accuracy": "awesome"})

        chk_tar_path = os.path.join(".keepsake/checkpoints", chk.id + ".tar.gz")
        wait(
            lambda: os.path.exists(chk_tar_path), timeout_seconds=5, sleep_seconds=0.01,
        )
        time.sleep(0.1)  # wait to finish writing

        # test with already existing checkpoint
        tmpdir = tmpdir_factory.mktemp("checkout")
        chk.checkout(output_directory=str(tmpdir))
        with open(tmpdir / "foo.txt") as f:
            assert f.read() == "foo"
        with open(tmpdir / "bar.txt") as f:
            assert f.read() == "bar"

        # test with checkpoint from keepsake.experiments.list()
        exp = project.experiments.list()[0]
        chk = exp.checkpoints[0]
        tmpdir = tmpdir_factory.mktemp("checkout")
        chk.checkout(output_directory=str(tmpdir))
        with open(tmpdir / "foo.txt") as f:
            assert f.read() == "foo"
        with open(tmpdir / "bar.txt") as f:
            assert f.read() == "bar"

        # test with no paths
        exp = project.experiments.create(params={"foo": "bar"}, disable_heartbeat=True)
        chk = exp.checkpoint(metrics={"accuracy": "awesome"})
        tmpdir = tmpdir_factory.mktemp("checkout")
        with pytest.raises(DoesNotExist):
            chk.checkout(output_directory=str(tmpdir))

        # test experiment with no path
        exp = project.experiments.create(params={"foo": "bar"}, disable_heartbeat=True)
        chk = exp.checkpoint(path="bar.txt", metrics={"accuracy": "awesome"})

        chk_tar_path = os.path.join(".keepsake/checkpoints", chk.id + ".tar.gz")
        wait(
            lambda: os.path.exists(chk_tar_path), timeout_seconds=5, sleep_seconds=0.01,
        )
        time.sleep(0.1)  # wait to finish writing

        tmpdir = tmpdir_factory.mktemp("checkout")
        chk.checkout(output_directory=str(tmpdir))
        assert not os.path.exists(tmpdir / "foo.txt")
        with open(tmpdir / "bar.txt") as f:
            assert f.read() == "bar"

        # test checkpoint with no path
        exp = project.experiments.create(
            path="foo.txt", params={"foo": "bar"}, disable_heartbeat=True
        )
        chk = exp.checkpoint(metrics={"accuracy": "awesome"})

        exp_tar_path = os.path.join(".keepsake/experiments", exp.id + ".tar.gz")
        wait(
            lambda: os.path.exists(exp_tar_path), timeout_seconds=5, sleep_seconds=0.01,
        )
        time.sleep(0.1)  # wait to finish writing

        tmpdir = tmpdir_factory.mktemp("checkout")
        chk.checkout(output_directory=str(tmpdir))
        assert not os.path.exists(tmpdir / "bar.txt")
        with open(tmpdir / "foo.txt") as f:
            assert f.read() == "foo"

    def test_open(self, temp_workdir):
        project = Project()
        with open("foo.txt", "w") as f:
            f.write("foo")

        with open("keepsake.yaml", "w") as f:
            f.write("repository: file://.keepsake/")

        exp = project.experiments.create(
            path=".", params={"foo": "bar"}, disable_heartbeat=True
        )
        with open("bar.txt", "w") as f:
            f.write("bar")
        chk = exp.checkpoint(path="bar.txt", metrics={"accuracy": "awesome"})

        chk_tar_path = os.path.join(".keepsake/checkpoints", chk.id + ".tar.gz")
        wait(
            lambda: os.path.exists(chk_tar_path), timeout_seconds=5, sleep_seconds=0.01,
        )
        time.sleep(0.1)  # wait to finish writing

        # test with already existing checkpoint
        assert chk.open("foo.txt").read().decode() == "foo"
        assert chk.open("bar.txt").read().decode() == "bar"

        # test with checkpoint from keepsake.experiments.list()
        exp = project.experiments.list()[0]
        chk = exp.checkpoints[0]
        assert chk.open("foo.txt").read().decode() == "foo"
        assert chk.open("bar.txt").read().decode() == "bar"


class TestCheckpointList:
    def test_metrics(self):
        experiment = experiment_factory(
            id="e1",
            checkpoints=[
                checkpoint_factory(id="c1", metrics={"loss": 0.1}),
                checkpoint_factory(id="c2"),
                checkpoint_factory(id="c3", metrics={"foo": "bar"}),
                checkpoint_factory(id="c3", metrics={"loss": 0.2}),
            ],
        )
        assert experiment.checkpoints.metrics["loss"] == [0.1, None, None, 0.2]

    def test_step(self):
        experiment = experiment_factory(
            id="e1",
            checkpoints=[
                checkpoint_factory(id="c1", step=10),
                checkpoint_factory(id="c2"),
                checkpoint_factory(id="c3", step=20),
            ],
        )
        assert experiment.checkpoints.step == [10, None, 20]

    def test_slice(self):
        experiment = experiment_factory(
            id="e1",
            checkpoints=[
                checkpoint_factory(id="c1"),
                checkpoint_factory(id="c2"),
                checkpoint_factory(id="c3"),
            ],
        )
        assert isinstance(experiment.checkpoints[:2], CheckpointList)
