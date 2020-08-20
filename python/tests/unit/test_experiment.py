import json
import os
import tempfile
import pytest  # type: ignore

import replicate


@pytest.fixture
def temp_workdir():
    orig_cwd = os.getcwd()
    try:
        with tempfile.TemporaryDirectory() as tmpdir:
            os.chdir(tmpdir)
            yield
    finally:
        os.chdir(orig_cwd)


def test_init_and_commit(temp_workdir):
    with open("train.py", "w") as fh:
        fh.write("print(1 + 1)")

    experiment = replicate.init(learning_rate=0.002)

    assert len(experiment.id) == 64
    with open(
        ".replicate/storage/metadata/experiments/{}.json".format(experiment.id)
    ) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == experiment.id
    assert metadata["params"] == {"learning_rate": 0.002}

    with open(".replicate/storage/experiments/{}/train.py".format(experiment.id)) as fh:
        assert fh.read() == "print(1 + 1)"

    # commit with a file
    with open("weights", "w") as fh:
        fh.write("1.2kg")

    commit = experiment.commit(path="weights", step=1, validation_loss=0.123)

    assert len(commit.id) == 64
    with open(".replicate/storage/metadata/commits/{}.json".format(commit.id)) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == commit.id
    assert metadata["step"] == 1
    assert metadata["labels"] == {"validation_loss": 0.123}
    assert metadata["experiment_id"] == experiment.id
    with open(".replicate/storage/commits/{}/weights".format(commit.id)) as fh:
        assert fh.read() == "1.2kg"
    assert (
        os.path.exists(".replicate/storage/commits/{}/train.py".format(commit.id))
        is False
    )

    # commit with a directory
    os.mkdir("data")
    with open("data/weights", "w") as fh:
        fh.write("1.3kg")

    commit = experiment.commit(path="data", step=1, validation_loss=0.123)

    with open(".replicate/storage/commits/{}/data/weights".format(commit.id)) as fh:
        assert fh.read() == "1.3kg"
    assert (
        os.path.exists(".replicate/storage/commits/{}/train.py".format(commit.id))
        is False
    )

    # commit with no path
    commit = experiment.commit(path=None, step=1, validation_loss=0.123)
    with open(".replicate/storage/metadata/commits/{}.json".format(commit.id)) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == commit.id
    assert os.path.exists(".replicate/storage/commits/{}".format(commit.id)) is False

    # commit requires path option
    with pytest.raises(TypeError, match="missing 1 required positional argument"):
        commit = experiment.commit()


def test_heartbeat(temp_workdir):
    experiment = replicate.init()
    assert experiment.heartbeat.path == "metadata/heartbeats/{}.json".format(
        experiment.id
    )
