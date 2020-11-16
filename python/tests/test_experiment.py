try:
    import dataclasses
except ImportError:
    from replicate._vendor import dataclasses
import datetime
import json
import os
import pytest  # type: ignore
import tarfile
import tempfile
from pathlib import Path
from unittest.mock import patch
from waiting import wait

import replicate
from replicate.exceptions import (
    DoesNotExistError,
    ConfigNotFoundError,
    NewerRepositoryVersion,
)
from replicate.experiment import Experiment, ExperimentList
from replicate.project import Project


from tests.factories import experiment_factory, checkpoint_factory


def test_init_and_checkpoint(temp_workdir):
    with open("replicate.yaml", "w") as f:
        f.write("repository: file://.replicate/")

    with open("train.py", "w") as fh:
        fh.write("print(1 + 1)")

    with open("README.md", "w") as fh:
        fh.write("Hello")

    # basic experiment
    experiment = replicate.init(
        path=".", params={"learning_rate": 0.002}, disable_heartbeat=True
    )

    assert len(experiment.id) == 64
    with open(".replicate/metadata/experiments/{}.json".format(experiment.id)) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == experiment.id
    assert metadata["params"] == {"learning_rate": 0.002}

    with tempfile.TemporaryDirectory() as tmpdir:
        with tarfile.open(
            ".replicate/experiments/{}.tar.gz".format(experiment.id)
        ) as tar:
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

    assert len(checkpoint.id) == 64
    with open(".replicate/metadata/experiments/{}.json".format(experiment.id)) as fh:
        metadata = json.load(fh)
    assert len(metadata["checkpoints"]) == 1
    checkpoint_metadata = metadata["checkpoints"][0]
    assert checkpoint_metadata["id"] == checkpoint.id
    assert checkpoint_metadata["step"] == 1
    assert checkpoint_metadata["metrics"] == {"validation_loss": 0.123}

    with tempfile.TemporaryDirectory() as tmpdir:
        with tarfile.open(
            ".replicate/checkpoints/{}.tar.gz".format(checkpoint.id)
        ) as tar:
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

    with tempfile.TemporaryDirectory() as tmpdir:
        with tarfile.open(
            ".replicate/checkpoints/{}.tar.gz".format(checkpoint.id)
        ) as tar:
            tar.extractall(tmpdir)

        assert (
            open(os.path.join(tmpdir, checkpoint.id, "data/weights")).read() == "1.3kg"
        )
        assert not os.path.exists(os.path.join(tmpdir, checkpoint.id, "train.py"))

    # checkpoint with no path
    checkpoint = experiment.checkpoint(
        path=None, step=1, metrics={"validation_loss": 0.123}
    )
    with open(".replicate/metadata/experiments/{}.json".format(experiment.id)) as fh:
        metadata = json.load(fh)
    assert metadata["checkpoints"][-1]["id"] == checkpoint.id
    assert not os.path.exists(".replicate/checkpoints/{}.tar.gz".format(checkpoint.id))

    # experiment with file
    experiment = replicate.init(
        path="train.py", params={"learning_rate": 0.002}, disable_heartbeat=True
    )
    with tempfile.TemporaryDirectory() as tmpdir:
        with tarfile.open(
            ".replicate/experiments/{}.tar.gz".format(experiment.id)
        ) as tar:
            tar.extractall(tmpdir)

        assert (
            open(os.path.join(tmpdir, experiment.id, "train.py")).read()
            == "print(1 + 1)"
        )
        assert not os.path.exists(os.path.join(tmpdir, experiment.id, "README.md"))

    # experiment with no path!
    experiment = replicate.init(
        path=None, params={"learning_rate": 0.002}, disable_heartbeat=True
    )
    with open(".replicate/metadata/experiments/{}.json".format(experiment.id)) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == experiment.id
    assert metadata["params"] == {"learning_rate": 0.002}
    assert not os.path.exists(".replicate/experiments/{}.tar.gz".format(experiment.id))


def test_init_with_config_file(temp_workdir):
    with open("replicate.yaml", "w") as f:
        f.write("repository: file://.replicate/")
    experiment = replicate.init()
    assert isinstance(experiment, Experiment)
    experiment.stop()


def test_init_without_config_file(temp_workdir):
    with pytest.raises(ConfigNotFoundError):
        replicate.init()


def test_heartbeat(temp_workdir):
    with open("replicate.yaml", "w") as f:
        f.write("repository: file://.replicate/")

    experiment = replicate.init()
    heartbeat_path = f".replicate/metadata/heartbeats/{experiment.id}.json"
    wait(lambda: os.path.exists(heartbeat_path), timeout_seconds=1, sleep_seconds=0.01)
    assert json.load(open(heartbeat_path))["experiment_id"] == experiment.id
    experiment.stop()
    assert not os.path.exists(heartbeat_path)

    # check starting and stopping immediately doesn't do anything weird
    experiment = replicate.init()
    experiment.stop()


def test_deprecated_repository_backwards_compatible(temp_workdir):
    os.makedirs(".replicate/storage")
    experiment = replicate.init()
    assert isinstance(experiment, Experiment)
    assert experiment._project._repository_url == "file://.replicate/storage"
    experiment.stop()

    with open("replicate.yaml", "w") as f:
        f.write("repository: file://foobar")
    experiment = replicate.init()
    assert isinstance(experiment, Experiment)
    assert experiment._project._repository_url == "file://foobar"
    experiment.stop()


def test_project_repository_version(temp_workdir):
    with open("replicate.yaml", "w") as f:
        f.write("repository: file://.replicate")
    experiment = replicate.init()

    expected = """{
  "version": 1
}"""
    with open(".replicate/repository.json") as f:
        assert f.read() == expected

    # no error on second init
    experiment = replicate.init()
    with open(".replicate/repository.json") as f:
        # repository.json shouldn't have changed
        assert f.read() == expected

    with open(".replicate/repository.json", "w") as f:
        f.write(
            """{
  "version": 2
}"""
        )
    with pytest.raises(NewerRepositoryVersion):
        replicate.init()


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

        experiment = Experiment(path="..", **kwargs)
        assert (
            "The path passed to the experiment must not start with '..' or '/'."
            in experiment.validate()[0]
        )
        experiment = Experiment(path="/", **kwargs)
        assert (
            "The path passed to the experiment must not start with '..' or '/'."
            in experiment.validate()[0]
        )
        experiment = Experiment(path="blah", **kwargs)
        assert (
            "The path passed to the experiment does not exist: blah"
            in experiment.validate()[0]
        )

    def test_from_json(self):
        data = {
            "id": "3132f9288bcc09a6b4d283c95a3968379d6b01fcf5d06500e789f90fdb02b7e1",
            "created": "2020-10-07T22:44:06.243914Z",
            "params": {"learning_rate": 0.01, "num_epochs": 100},
            "user": "ben",
            "host": "",
            "command": "train.py",
            "config": {"repository": ".replicate/"},
            "path": ".",
            "python_packages": {"foo": "1.0.0"},
            "checkpoints": [],
            "replicate_version": "0.0.1",
        }
        exp = Experiment.from_json(None, data)
        assert dataclasses.asdict(exp) == {
            "id": "3132f9288bcc09a6b4d283c95a3968379d6b01fcf5d06500e789f90fdb02b7e1",
            "created": datetime.datetime(2020, 10, 7, 22, 44, 6, 243914),
            "params": {"learning_rate": 0.01, "num_epochs": 100},
            "user": "ben",
            "host": "",
            "command": "train.py",
            "config": {"repository": ".replicate/"},
            "path": ".",
            "python_packages": {"foo": "1.0.0"},
            "checkpoints": [],
            "replicate_version": "0.0.1",
        }

    def test_checkpoints(self, temp_workdir):
        project = Project()

        with open("replicate.yaml", "w") as f:
            f.write("repository: file://.replicate/")

        experiment = project.experiments.create(
            path=None, params={"foo": "bar"}, disable_heartbeat=True
        )
        chk1 = experiment.checkpoint(path=None, metrics={"accuracy": "ok"})
        chk2 = experiment.checkpoint(path=None, metrics={"accuracy": "super"})
        assert len(experiment.checkpoints) == 2
        assert experiment.checkpoints[0].id == chk1.id
        assert experiment.checkpoints[1].id == chk2.id

    def test_delete(self, temp_workdir):
        project = Project()

        with open("replicate.yaml", "w") as f:
            f.write("repository: file://.replicate/")

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
                str(p).replace(".replicate/", "") for p in Path(".replicate").rglob("*")
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


class TestExperimentCollection:
    def test_get(self, temp_workdir):
        project = Project()

        with open("replicate.yaml", "w") as f:
            f.write("repository: file://.replicate/")

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

        with pytest.raises(DoesNotExistError):
            project.experiments.get("doesnotexist")

    def test_list(self, temp_workdir):
        project = Project()

        with open("replicate.yaml", "w") as f:
            f.write("repository: file://.replicate/")

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
            (False, False, False, ConfigNotFoundError),

            # has config -> good
            (False, False, True, None),

            # has directory but no repo -> bad
            (False, True, False, ConfigNotFoundError),

            # has directory but no repo, and config exists -> good
            (False, True, True, None),

            # has repo but no directory -> bad
            (True, False, False, ValueError),
            (True, False, True, ValueError),  # even with config

            # has repo and directory -> good
            (True, True, False, None),
            (True, True, True, None),  # even with config
        ],
        # fmt: on
    )
    def test_create_project_options(
        self, has_repository, has_directory, has_config, exception, temp_workdir
    ):
        repo = "file://.replicate/" if has_repository else None
        directory = "." if has_directory else None

        if has_config:
            with open("replicate.yaml", "w") as f:
                f.write("repository: file://.replicate/")

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
        repo = "file://.replicate/" if has_repository else None
        directory = "." if has_directory else None

        if has_config:
            with open("replicate.yaml", "w") as f:
                f.write("repository: file://.replicate/")

        project = Project(repository=repo, directory=directory)
        if should_error:
            with pytest.raises((ValueError, ConfigNotFoundError)):
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
