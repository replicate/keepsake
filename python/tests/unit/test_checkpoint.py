import replicate
from replicate.checkpoint import Checkpoint


class Blah:
    pass


def test_validate():
    experiment = replicate.init(
        path=None, params={"foo": "bar"}, disable_heartbeat=True
    )

    checkpoint = Checkpoint(experiment, path=123)
    assert checkpoint.validate() == ["path must be a string"]

    checkpoint = Checkpoint(experiment, step="lol")
    assert checkpoint.validate() == ["step must be an integer"]

    checkpoint = Checkpoint(experiment, metrics="lol")
    assert checkpoint.validate() == ["metrics must be a dictionary"]

    checkpoint = Checkpoint(experiment, metrics={"foo": Blah()})
    assert "Failed to serialize the metric 'foo' to JSON" in checkpoint.validate()[0]

    checkpoint = Checkpoint(
        experiment,
        metrics={"foo": "bar"},
        primary_metric_name="baz",
        primary_metric_goal="maximize",
    )
    assert checkpoint.validate() == ["Primary metric 'baz' is not defined in metrics"]

    checkpoint = Checkpoint(
        experiment,
        metrics={"foo": "bar"},
        primary_metric_name="foo",
        primary_metric_goal="maximilize",
    )
    assert (
        "Primary metric goal must be either 'maximize' or 'minimize'"
        in checkpoint.validate()[0]
    )
