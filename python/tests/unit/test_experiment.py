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


def test_init(temp_workdir):
    experiment = replicate.init(params={"learning_rate": 0.002})

    assert len(experiment.id) == 64
    with open(
        ".replicate/storage/experiments/{}/replicate-metadata.json".format(
            experiment.id
        )
    ) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == experiment.id
    assert metadata["params"] == {"learning_rate": 0.002}


def test_commit(temp_workdir):
    with open("train.py", "w") as fh:
        fh.write("print(1 + 1)")

    experiment = replicate.init(params={"learning_rate": 0.002})
    commit = experiment.commit({"validation_loss": 0.123})

    assert len(commit.id) == 64
    with open(
        ".replicate/storage/commits/{}/replicate-metadata.json".format(commit.id)
    ) as fh:
        metadata = json.load(fh)
    assert metadata["id"] == commit.id
    assert metadata["metrics"] == {"validation_loss": 0.123}
    assert metadata["experiment"]["id"] == experiment.id

    with open(".replicate/storage/commits/{}/train.py".format(commit.id)) as fh:
        assert fh.read() == "print(1 + 1)"
