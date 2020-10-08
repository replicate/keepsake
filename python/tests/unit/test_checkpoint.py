import datetime

from replicate.checkpoint import Checkpoint
from replicate.experiment import Experiment


class Blah:
    pass


def test_validate():
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
    assert "Failed to serialize the metric 'foo' to JSON" in checkpoint.validate()[0]

    checkpoint = Checkpoint(
        metrics={"foo": "bar"},
        primary_metric={"name": "baz", "goal": "maximize"},
        **kwargs
    )
    assert checkpoint.validate() == ["Primary metric 'baz' is not defined in metrics"]

    checkpoint = Checkpoint(
        metrics={"foo": "bar"},
        primary_metric={"name": "foo", "goal": "maximilize"},
        **kwargs
    )
    assert (
        "Primary metric goal must be either 'maximize' or 'minimize'"
        in checkpoint.validate()[0]
    )
