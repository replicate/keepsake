try:
    import dataclasses
except ImportError:
    from keepsake._vendor import dataclasses
import math
import datetime
import json
import os
import pytest  # type: ignore
import tarfile
import tempfile
import time
from pathlib import Path
from unittest.mock import patch
from waiting import wait

import keepsake
from keepsake.exceptions import (
    DoesNotExist,
    ConfigNotFound,
    IncompatibleRepositoryVersion,
)
from keepsake.experiment import Experiment, ExperimentList
from keepsake.project import Project
from keepsake.metadata import rfc3339_datetime

from tests.factories import experiment_factory, checkpoint_factory


def test_init_and_checkpoint(temp_workdir):
    with open("keepsake.yaml", "w") as f:
        f.write("repository: file://.keepsake/")

    with open("train.py", "w") as fh:
        fh.write("print(1 + 1)")

    with open("README.md", "w") as fh:
        fh.write("Hello")

    # basic experiment
    experiment = keepsake.init(
        path=".", params={"learning_rate": 0.002}, disable_heartbeat=True
    )

    experiment_tar_path = ".keepsake/experiments/{}.tar.gz".format(experiment.id)
    wait(
        lambda: os.path.exists(experiment_tar_path),
        timeout_seconds=5,
        sleep_seconds=0.01,
    )
    time.sleep(0.1)  # wait for file to be written

    assert len(experiment.id) == 64
    with open(".keepsake/metadata/experiments/{}.json".format(experiment.id)) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == experiment.id
    assert metadata["params"] == {"learning_rate": 0.002}
    assert metadata["host"] == ""
    assert metadata["user"] != ""
    # FIXME: this is broken https://github.com/replicate/keepsake/issues/492
    assert metadata["config"]["repository"].startswith("file://")
    assert metadata["command"] != ""
    assert metadata["path"] == "."
    assert metadata["python_version"] != ""
    assert len(metadata["python_packages"]) > 0
    assert metadata["keepsake_version"] != ""

    with tempfile.TemporaryDirectory() as tmpdir:
        with tarfile.open(experiment_tar_path) as tar:
            tar.extractall(tmpdir)

        assert (
            open(os.path.join(tmpdir, experiment.id, "train.py")).read()
            == "print(1 + 1)"
        )
        assert os.path.exists(os.path.join(tmpdir, experiment.id, "README.md"))

    # checkpoint with a file
    with open("weights", "w") as fh:
        fh.write("1.2kg")

    checkpoint = experiment.checkpoint(
        path="weights", step=1, metrics={"validation_loss": 0.123}
    )

    checkpoint_tar_path = ".keepsake/checkpoints/{}.tar.gz".format(checkpoint.id)
    wait(
        lambda: os.path.exists(checkpoint_tar_path),
        timeout_seconds=5,
        sleep_seconds=0.01,
    )
    time.sleep(0.1)  # wait for file to be written

    assert len(checkpoint.id) == 64
    with open(".keepsake/metadata/experiments/{}.json".format(experiment.id)) as fh:
        metadata = json.load(fh)
    assert len(metadata["checkpoints"]) == 1
    checkpoint_metadata = metadata["checkpoints"][0]
    assert checkpoint_metadata["id"] == checkpoint.id
    assert checkpoint_metadata["step"] == 1
    assert checkpoint_metadata["metrics"] == {"validation_loss": 0.123}

    with tempfile.TemporaryDirectory() as tmpdir:
        with tarfile.open(checkpoint_tar_path) as tar:
            tar.extractall(tmpdir)

        assert open(os.path.join(tmpdir, checkpoint.id, "weights")).read() == "1.2kg"
        assert not os.path.exists(os.path.join(tmpdir, checkpoint.id, "train.py"))

    # checkpoint with a directory
    os.mkdir("data")
    with open("data/weights", "w") as fh:
        fh.write("1.3kg")

    checkpoint = experiment.checkpoint(
        path="data", step=1, metrics={"validation_loss": 0.123}
    )

    checkpoint_tar_path = ".keepsake/checkpoints/{}.tar.gz".format(checkpoint.id)
    wait(
        lambda: os.path.exists(checkpoint_tar_path),
        timeout_seconds=5,
        sleep_seconds=0.01,
    )
    time.sleep(0.1)  # wait for file to be written

    with tempfile.TemporaryDirectory() as tmpdir:
        with tarfile.open(checkpoint_tar_path) as tar:
            tar.extractall(tmpdir)

        assert (
            open(os.path.join(tmpdir, checkpoint.id, "data/weights")).read() == "1.3kg"
        )
        assert not os.path.exists(os.path.join(tmpdir, checkpoint.id, "train.py"))

    # checkpoint with no path
    checkpoint = experiment.checkpoint(
        path=None, step=1, metrics={"validation_loss": 0.123}
    )

    # wait in case async process tries to create a path anyway
    time.sleep(0.5)

    with open(".keepsake/metadata/experiments/{}.json".format(experiment.id)) as fh:
        metadata = json.load(fh)
    assert metadata["checkpoints"][-1]["id"] == checkpoint.id
    assert not os.path.exists(".keepsake/checkpoints/{}.tar.gz".format(checkpoint.id))

    # experiment with file
    experiment = keepsake.init(
        path="train.py", params={"learning_rate": 0.002}, disable_heartbeat=True
    )

    experiment_tar_path = ".keepsake/experiments/{}.tar.gz".format(experiment.id)
    wait(
        lambda: os.path.exists(experiment_tar_path),
        timeout_seconds=5,
        sleep_seconds=0.01,
    )
    time.sleep(0.1)  # wait for file to be written

    with tempfile.TemporaryDirectory() as tmpdir:
        with tarfile.open(experiment_tar_path) as tar:
            tar.extractall(tmpdir)

        assert (
            open(os.path.join(tmpdir, experiment.id, "train.py")).read()
            == "print(1 + 1)"
        )
        assert not os.path.exists(os.path.join(tmpdir, experiment.id, "README.md"))

    # experiment with no path!
    experiment = keepsake.init(
        path=None, params={"learning_rate": 0.002}, disable_heartbeat=True
    )

    # wait in case async process tries to create a path anyway
    time.sleep(0.5)

    with open(".keepsake/metadata/experiments/{}.json".format(experiment.id)) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == experiment.id
    assert metadata["params"] == {"learning_rate": 0.002}
    assert not os.path.exists(".keepsake/experiments/{}.tar.gz".format(experiment.id))


def test_init_with_config_file(temp_workdir):
    with open("keepsake.yaml", "w") as f:
        f.write("repository: file://.keepsake/")
    experiment = keepsake.init()
    assert isinstance(experiment, Experiment)
    experiment.stop()


def test_init_without_config_file(temp_workdir):
    with pytest.raises(ConfigNotFound):
        keepsake.init()


def test_project_repository_version(temp_workdir):
    with open("keepsake.yaml", "w") as f:
        f.write("repository: file://.keepsake")
    experiment = keepsake.init()

    expected = """{"version":1}"""
    with open(".keepsake/repository.json") as f:
        assert f.read() == expected

    # no error on second init
    experiment = keepsake.init()
    with open(".keepsake/repository.json") as f:
        # repository.json shouldn't have changed
        assert f.read() == expected

    with open(".keepsake/repository.json", "w") as f:
        f.write("""{"version":2}""")
    with pytest.raises(IncompatibleRepositoryVersion):
        keepsake.init()


def test_is_running(temp_workdir):
    with open("keepsake.yaml", "w") as f:
        f.write("repository: file://.keepsake/")

    experiment = keepsake.init()

    heartbeat_path = f".keepsake/metadata/heartbeats/{experiment.id}.json"

    assert wait(
        lambda: os.path.exists(heartbeat_path), timeout_seconds=10, sleep_seconds=0.01
    )

    # Check whether experiment is running after heartbeats are started
    assert experiment.is_running()

    # Heartbeats stopped
    experiment.stop()
    assert not experiment.is_running()


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
        assert experiment._validate() == ["params must be a dictionary"]

        experiment = Experiment(path=None, params={"foo": Blah()}, **kwargs)
        assert (
            "Failed to serialize the param 'foo' to JSON" in experiment._validate()[0]
        )

        experiment = Experiment(path="..", **kwargs)
        assert (
            "The path passed to the experiment must not start with '..' or '/'."
            in experiment._validate()[0]
        )
        experiment = Experiment(path="/", **kwargs)
        assert (
            "The path passed to the experiment must not start with '..' or '/'."
            in experiment._validate()[0]
        )
        experiment = Experiment(path="blah", **kwargs)
        assert (
            "The path passed to the experiment does not exist: blah"
            in experiment._validate()[0]
        )

    def test_checkpoints(self, temp_workdir):
        project = Project()

        with open("keepsake.yaml", "w") as f:
            f.write("repository: file://.keepsake/")

        experiment = project.experiments.create(
            path=None, params={"foo": "bar"}, disable_heartbeat=True
        )
        chk1 = experiment.checkpoint(path=None, metrics={"accuracy": "ok"})
        chk2 = experiment.checkpoint(path=None, metrics={"accuracy": "super"})
        assert len(experiment.checkpoints) == 2
        assert experiment.checkpoints[0].id == chk1.id
        assert experiment.checkpoints[1].id == chk2.id

    def test_checkpoint_auto_increments_step(self, temp_workdir):
        project = Project()

        with open("keepsake.yaml", "w") as f:
            f.write("repository: file://.keepsake/")

        experiment = project.experiments.create(
            path=None, params={"foo": "bar"}, disable_heartbeat=True
        )
        chk1 = experiment.checkpoint()
        chk2 = experiment.checkpoint()
        chk3 = experiment.checkpoint(step=10)
        chk4 = experiment.checkpoint()
        assert chk1.step == 0
        assert chk2.step == 1
        assert chk3.step == 10
        assert chk4.step == 11

    def test_delete(self, temp_workdir):
        project = Project()

        with open("keepsake.yaml", "w") as f:
            f.write("repository: file://.keepsake/")

        with open("foo.txt", "w") as f:
            f.write("hello")

        experiment = project.experiments.create(
            path=".", params={"foo": "bar"}, disable_heartbeat=True
        )
        with open("model.txt", "w") as f:
            f.write("i'm a model")
        chk = experiment.checkpoint(path="model.txt", metrics={"accuracy": "awesome"})

        def get_paths():
            return set(
                str(p).replace(".keepsake/", "") for p in Path(".keepsake").rglob("*")
            )

        chk_tar_path = os.path.join(".keepsake/checkpoints", chk.id + ".tar.gz")
        wait(
            lambda: os.path.exists(chk_tar_path), timeout_seconds=5, sleep_seconds=0.01,
        )

        paths = get_paths()
        expected = set(
            [
                "repository.json",
                "metadata/experiments/{}.json".format(experiment.id),
                "experiments",
                "checkpoints/{}.tar.gz".format(chk.id),
                "metadata",
                "metadata/experiments",
                "experiments/{}.tar.gz".format(experiment.id),
                "checkpoints",
            ]
        )
        assert paths == expected

        experiment.delete()

        paths = get_paths()
        expected = set(
            [
                "repository.json",  # we're not deleting the project spec
                "experiments",
                "metadata",
                "metadata/experiments",
                "checkpoints",
            ]
        )
        assert paths == expected

    def test_refresh(self, temp_workdir):
        project = Project()

        with open("keepsake.yaml", "w") as f:
            f.write("repository: file://.keepsake/")

        experiment = project.experiments.create(
            params={"foo": "bar"}, disable_heartbeat=True
        )

        experiment.checkpoint(metrics={"accuracy": 0})

        other_experiment = project.experiments.get(experiment.id)
        assert len(other_experiment.checkpoints) == 1

        experiment.checkpoint(metrics={"accuracy": 1})
        assert len(other_experiment.checkpoints) == 1

        other_experiment.refresh()
        assert len(other_experiment.checkpoints) == 2
        assert other_experiment.checkpoints[-1].metrics["accuracy"] == 1

    def test_best_none(self, temp_workdir):
        project = Project()

        with open("keepsake.yaml", "w") as f:
            f.write("repository: file://.keepsake/")

        experiment = project.experiments.create(disable_heartbeat=True)

        experiment.checkpoint(
            path=None,
            metrics={"accuracy": None},
            primary_metric=("accuracy", "maximize"),
        )
        experiment.checkpoint(
            path=None,
            metrics={"accuracy": float("nan")},
            primary_metric=("accuracy", "maximize"),
        )
        assert experiment.best() is None

    def test_exceptional_values(self, temp_workdir):
        project = Project()

        with open("keepsake.yaml", "w") as f:
            f.write("repository: file://.keepsake/")

        experiment = project.experiments.create(disable_heartbeat=True)
        experiment.checkpoint(
            path=None,
            metrics={"accuracy": float("nan")},
            primary_metric=("accuracy", "maximize"),
        )
        experiment.checkpoint(
            path=None,
            metrics={"accuracy": float("-inf")},
            primary_metric=("accuracy", "maximize"),
        )
        experiment.checkpoint(
            path=None,
            metrics={"accuracy": float("+inf")},
            primary_metric=("accuracy", "maximize"),
        )
        experiment.checkpoint(
            path=None,
            metrics={"accuracy": None},
            primary_metric=("accuracy", "maximize"),
        )

        experiment = project.experiments.get(experiment.id)
        assert math.isnan(experiment.checkpoints[0].metrics["accuracy"])
        assert math.isinf(experiment.checkpoints[1].metrics["accuracy"])
        assert experiment.checkpoints[1].metrics["accuracy"] < 0
        assert math.isinf(experiment.checkpoints[2].metrics["accuracy"])
        assert experiment.checkpoints[2].metrics["accuracy"] > 0
        assert experiment.checkpoints[3].metrics["accuracy"] is None


class TestExperimentCollection:
    def test_get(self, temp_workdir):
        project = Project()

        with open("keepsake.yaml", "w") as f:
            f.write("repository: file://.keepsake/")

        exp1 = project.experiments.create(
            path=None, params={"foo": "bar"}, disable_heartbeat=True
        )
        exp1.checkpoint(path=None, metrics={"accuracy": "wicked"})
        exp2 = project.experiments.create(
            path=None, params={"foo": "baz"}, disable_heartbeat=True
        )

        actual_exp = project.experiments.get(exp1.id)
        assert actual_exp.created == exp1.created
        assert len(actual_exp.checkpoints) == 1
        assert actual_exp.checkpoints[0].metrics == {"accuracy": "wicked"}
        # get by prefix
        assert project.experiments.get(exp2.id[:7]).created == exp2.created

        with pytest.raises(DoesNotExist):
            project.experiments.get("doesnotexist")

    def test_list(self, temp_workdir):
        project = Project()

        with open("keepsake.yaml", "w") as f:
            f.write("repository: file://.keepsake/")

        exp1 = project.experiments.create(
            path=None, params={"foo": "bar"}, disable_heartbeat=True
        )
        exp1.checkpoint(path=None, metrics={"accuracy": "wicked"})
        exp2 = project.experiments.create(
            path=None, params={"foo": "baz"}, disable_heartbeat=True
        )

        experiments = project.experiments.list()
        assert len(experiments) == 2
        assert experiments[0].id == exp1.id
        assert len(experiments[0].checkpoints) == 1
        assert experiments[0].checkpoints[0].metrics == {"accuracy": "wicked"}
        assert experiments[1].id == exp2.id

    # FIXME(bfirsh): these parameterized tests are hard to parse as a reader -- might be better written out verbosely as code?
    @pytest.mark.parametrize(
        "has_repository,has_directory,has_config,exception",
        # fmt: off
        [
            # nothing -> bad
            (False, False, False, ConfigNotFound),

            # has config -> good
            (False, False, True, None),

            # has directory but no repo -> bad
            (False, True, False, ConfigNotFound),

            # has directory but no repo, and config exists -> good
            (False, True, True, None),

            # has repo but no directory, uses current working directory by default -> good
            (True, False, False, None),

            # has repo, no directory, but infers directory from config -> good
            (True, False, True, None),

            # has repo and directory -> good
            (True, True, False, None),
            (True, True, True, None),  # even with config
        ],
        # fmt: on
    )
    def test_create_project_options(
        self, has_repository, has_directory, has_config, exception, temp_workdir
    ):
        repo = "file://.keepsake/" if has_repository else None
        directory = "." if has_directory else None

        if has_config:
            with open("keepsake.yaml", "w") as f:
                f.write("repository: file://.keepsake/")

        with open("foo.txt", "w") as f:
            f.write("hello world")

        project = Project(repository=repo, directory=directory)

        if exception:
            with pytest.raises(exception):
                project.experiments.create(path=".")

        else:
            exp = project.experiments.create(path=".")
            # to avoid writing heartbeats that sometimes cause
            # TemporaryDirectory cleanup to fail
            exp.stop()

    @pytest.mark.parametrize(
        "has_repository,has_directory,has_config,should_error",
        # fmt: off
        [
            # nothing -> bad
            (False, False, False, True),

            # has config -> good
            (False, False, True, False),

            # has directory but no repo -> bad
            (False, True, False, True),

            # has directory but no repo, and config exists -> good
            (False, True, True, False),

            # has repo but no directory -> GOOD (differs from create)
            (True, False, False, False),
            (True, False, True, False),  # even with config

            # has repo and directory -> good
            (True, True, False, False),
            (True, True, True, False),  # even with config
        ],
        # fmt: on
    )
    def test_list_project_options(
        self, has_repository, has_directory, has_config, should_error, temp_workdir
    ):
        repo = "file://.keepsake/" if has_repository else None
        directory = "." if has_directory else None

        if has_config:
            with open("keepsake.yaml", "w") as f:
                f.write("repository: file://.keepsake/")

        project = Project(repository=repo, directory=directory)
        if should_error:
            with pytest.raises((ValueError, ConfigNotFound)):
                project.experiments.list()
        else:
            exps = project.experiments.list()
            assert isinstance(exps, ExperimentList)
            assert len(exps) == 0


class TestExperimentList:
    def test_repr_html(self, temp_workdir):

        experiment_list = ExperimentList(
            [
                experiment_factory(
                    id="e1",
                    checkpoints=[
                        checkpoint_factory(
                            id="c1",
                            metrics={"loss": 0.1},
                            primary_metric={"name": "loss", "goal": "minimize"},
                        ),
                        checkpoint_factory(
                            id="c2",
                            metrics={"loss": 0.2},
                            primary_metric={"name": "loss", "goal": "minimize"},
                        ),
                    ],
                ),
                experiment_factory(
                    id="e2",
                    checkpoints=[
                        checkpoint_factory(
                            id="c3",
                            metrics={"loss": 0.2},
                            primary_metric={"name": "loss", "goal": "minimize"},
                        ),
                        checkpoint_factory(
                            id="c4",
                            metrics={"loss": 0.1},
                            primary_metric={"name": "loss", "goal": "minimize"},
                        ),
                    ],
                ),
            ]
        )

        assert (
            experiment_list._repr_html_()
            == """
<table><tr><th>id</th><th>created</th><th>params</th><th>latest_checkpoint</th><th>best_checkpoint</th></tr>
<tr><th>e1</th><th>2020-01-01 01:01:01</th><th>None</th><th>c2 (loss: 0.2)</th><th>c1 (loss: 0.1)</th></tr>
<tr><th>e2</th><th>2020-01-01 01:01:01</th><th>None</th><th>c4 (loss: 0.1)</th><th>c4 (loss: 0.1)</th></tr></table>""".strip().replace(
                "\n", ""
            )
        )
