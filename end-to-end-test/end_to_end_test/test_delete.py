import json
import os
import subprocess
from pathlib import Path
import pytest  # type: ignore

from .utils import path_exists


@pytest.mark.parametrize(
    "storage_backend,use_root",
    [
        ("gcs", False),
        ("gcs", True),
        ("s3", False),
        ("s3", True),
        pytest.param("file", False, marks=pytest.mark.fast),
    ],
)
def test_delete(storage_backend, use_root, tmpdir, temp_bucket, tmpdir_factory):
    tmpdir = str(tmpdir)
    if storage_backend == "s3":
        storage = "s3://" + temp_bucket
    if storage_backend == "gcs":
        storage = "gs://" + temp_bucket
    elif storage_backend == "file":
        storage = "file://" + str(tmpdir_factory.mktemp("storage"))

    # different root directory in buckets
    if use_root:
        storage += "/root"

    with open(Path(tmpdir) / "replicate.yaml", "w") as f:
        f.write(
            """
storage: {storage}
""".format(
                storage=storage
            )
        )
    with open(Path(tmpdir) / "train.py", "w") as f:
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

    env = os.environ
    env["PATH"] = "/usr/local/bin:" + os.environ["PATH"]

    cmd = ["python", "train.py", "--foo"]
    subprocess.run(cmd, cwd=tmpdir, env=env, check=True)

    experiments = json.loads(
        subprocess.run(
            ["replicate", "list", "--json"],
            cwd=tmpdir,
            env=env,
            capture_output=True,
            check=True,
        ).stdout
    )
    assert len(experiments) == 1
    assert experiments[0]["num_checkpoints"] == 3

    checkpoint_id = experiments[0]["latest_checkpoint"]["id"]
    checkpoint_storage_path = Path("checkpoints") / (checkpoint_id + ".tar.gz")
    assert path_exists(storage, checkpoint_storage_path)

    subprocess.run(
        ["replicate", "delete", checkpoint_id], cwd=tmpdir, env=env, check=True
    )

    experiments = json.loads(
        subprocess.run(
            ["replicate", "list", "--json"],
            cwd=tmpdir,
            env=env,
            capture_output=True,
            check=True,
        ).stdout
    )
    assert len(experiments) == 1
    # TODO(bfirsh): checkpoint metadata is no longer deleted, so check that checkout fails
    assert experiments[0]["num_checkpoints"] == 3
    assert not path_exists(storage, checkpoint_storage_path)

    experiment_id = experiments[0]["id"]
    experiment_storage_path = Path("experiments") / (experiment_id + ".tar.gz")

    assert path_exists(storage, experiment_storage_path)

    subprocess.run(
        ["replicate", "delete", experiment_id], cwd=tmpdir, env=env, check=True
    )

    experiments = json.loads(
        subprocess.run(
            ["replicate", "list", "--json"],
            cwd=tmpdir,
            env=env,
            capture_output=True,
            check=True,
        ).stdout
    )
    assert len(experiments) == 0
    assert not path_exists(storage, experiment_storage_path)
