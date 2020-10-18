try:
    import dataclasses
except ImportError:
    from replicate._vendor import dataclasses
import datetime

from replicate.checkpoint import Checkpoint


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
