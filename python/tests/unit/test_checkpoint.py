try:
    import dataclasses
except ImportError:
    from replicate._vendor import dataclasses
import datetime

from replicate.checkpoint import Checkpoint
from replicate.experiment import Experiment
from replicate.project import Project


class Blah:
    pass


class TestCheckpoint:
    def test_validate(self):
        experiment = Experiment(
            project=None,
            id="abc123",
            created=datetime.datetime.utcnow(),
            user="ben",
            host="",
            command="",
            config={},
            path=None,
            params={"foo": "bar"},
        )

        kwargs = {
            "experiment": experiment,
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

    def test_from_json(self, temp_workdir):
        data = {
            "id": "a1a973fcbead997a3b253c960b9fad1bb1386681beefd7ba8703e25548beb11c",
            "created": "2020-10-07T22:44:06.532785Z",
            "experiment_id": "3132f9288bcc09a6b4d283c95a3968379d6b01fcf5d06500e789f90fdb02b7e1",
            "path": "model.pth",
            "metrics": {"loss": 0.9042219519615173, "accuracy": 0.8666666746139526},
            "primary_metric": {"name": "loss", "goal": "minimize"},
            "step": 7,
        }
        project = Project()
        experiment = project.experiments.create()
        checkpoint = Checkpoint.from_json(experiment, data)
        assert dataclasses.asdict(checkpoint) == {
            "id": "a1a973fcbead997a3b253c960b9fad1bb1386681beefd7ba8703e25548beb11c",
            "created": datetime.datetime(2020, 10, 7, 22, 44, 6, 532785),
            "experiment": dataclasses.asdict(experiment),
            "path": "model.pth",
            "metrics": {"loss": 0.9042219519615173, "accuracy": 0.8666666746139526},
            "primary_metric": {"name": "loss", "goal": "minimize"},
            "step": 7,
        }


class TestCheckpointCollection:
    def test_list(self, temp_workdir):
        project = Project()
        experiment = project.experiments.create(path=None, params={"foo": "bar"})
        chk1 = experiment.checkpoints.create(path=None, metrics={"accuracy": "ok"})
        chk2 = experiment.checkpoints.create(path=None, metrics={"accuracy": "super"})
        checkpoints = experiment.checkpoints.list()
        assert len(checkpoints) == 2
        assert checkpoints[0].id == chk1.id
        assert checkpoints[1].id == chk2.id
