import json
import os
import subprocess
import pytest  # type: ignore

from .utils import get_env


@pytest.mark.parametrize(
    "storage_backend,use_root",
    [
        ("undefined", False),
        ("file", False),
        pytest.param("gcs", False, marks=pytest.mark.external),
        pytest.param("gcs", True, marks=pytest.mark.external),
        pytest.param("s3", False, marks=pytest.mark.external),
        pytest.param("s3", True, marks=pytest.mark.external),
    ],
)
def test_list(storage_backend, use_root, tmpdir, temp_bucket, tmpdir_factory):
    tmpdir = str(tmpdir)
    if storage_backend == "s3":
        storage = "s3://" + temp_bucket
    if storage_backend == "gcs":
        storage = "gs://" + temp_bucket
    elif storage_backend == "file":
        storage = "file://" + str(tmpdir_factory.mktemp("storage"))
    elif storage_backend == "undefined":
        storage = str(tmpdir_factory.mktemp("storage"))

    # different root directory in buckets
    if use_root:
        storage += "/root"

    with open(os.path.join(tmpdir, "replicate.yaml"), "w") as f:
        f.write(
            """
storage: {storage}
""".format(
                storage=storage
            )
        )
    with open(os.path.join(tmpdir, "train.py"), "w") as f:
        f.write(
            """
import replicate

def main():
    experiment = replicate.init(path=".", params={"my_param": "my-value"})

    for step in range(3):
        experiment.checkpoint(path=".", step=step)

if __name__ == "__main__":
    main()
"""
        )

    env = get_env()

    subprocess.run(["python", "train.py", "--foo"], cwd=tmpdir, env=env, check=True)

    experiments = json.loads(
        subprocess.run(
            ["replicate", "--verbose", "list", "--json"],
            cwd=tmpdir,
            env=env,
            stdout=subprocess.PIPE,
            check=True,
        ).stdout
    )

    assert len(experiments) == 1

    exp = experiments[0]
    assert len(exp["id"]) == 64
    assert exp["params"] == {"my_param": "my-value"}
    assert exp["num_checkpoints"] == 3
    assert exp["command"] == "train.py --foo"
    latest = exp["latest_checkpoint"]
    assert len(latest["id"]) == 64
    # FIXME: now rfc3339 strings
    # assert latest["created"] > exp["created"]
    assert latest["step"] == 2

    # test that --storage-url works
    experiments2 = json.loads(
        subprocess.run(
            ["replicate", "--verbose", "ls", "--json", "--storage-url=" + storage],
            cwd=tmpdir_factory.mktemp("list"),
            env=env,
            stdout=subprocess.PIPE,
            check=True,
        ).stdout
    )
    assert experiments2 == experiments

    # test incremental updates
    with open(os.path.join(tmpdir, "train2.py"), "w") as f:
        f.write(
            """
from replicate.project import Project
import sys

experiment_id = sys.argv[1]
experiment = Project().experiments.get(experiment_id)
experiment.checkpoint(path=".", step=3)
"""
        )
    subprocess.run(["python", "train2.py", exp["id"]], cwd=tmpdir, env=env, check=True)
    experiments = json.loads(
        subprocess.run(
            ["replicate", "--verbose", "list", "--json"],
            cwd=tmpdir,
            env=env,
            stdout=subprocess.PIPE,
            check=True,
        ).stdout
    )
    assert len(experiments) == 1
    exp = experiments[0]
    latest = exp["latest_checkpoint"]
    assert latest["step"] == 3
