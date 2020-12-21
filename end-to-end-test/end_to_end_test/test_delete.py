import json
import os
import subprocess
from pathlib import Path
import pytest  # type: ignore

from .utils import path_exists, get_env


@pytest.mark.parametrize(
    "repository_backend,use_root",
    [
        ("file", False),
        pytest.param("gcs", False, marks=pytest.mark.external),
        pytest.param("gcs", True, marks=pytest.mark.external),
        pytest.param("s3", False, marks=pytest.mark.external),
        pytest.param("s3", True, marks=pytest.mark.external),
    ],
)
def test_delete(
    repository_backend, use_root, tmpdir, temp_bucket_factory, tmpdir_factory
):
    tmpdir = str(tmpdir)
    if repository_backend == "s3":
        repository = "s3://" + temp_bucket_factory.s3()
    if repository_backend == "gcs":
        repository = "gs://" + temp_bucket_factory.gs()
    elif repository_backend == "file":
        repository = "file://" + str(tmpdir_factory.mktemp("repository"))

    # different root directory in buckets
    if use_root:
        repository += "/root"

    with open(Path(tmpdir) / "replicate.yaml", "w") as f:
        f.write(
            """
repository: {repository}
""".format(
                repository=repository
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

    env = get_env()
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
    assert path_exists(repository, checkpoint_storage_path)

    subprocess.run(
        ["replicate", "delete", "--force", checkpoint_id],
        cwd=tmpdir,
        env=env,
        check=True,
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
    assert not path_exists(repository, checkpoint_storage_path)

    experiment_id = experiments[0]["id"]
    experiment_storage_path = Path("experiments") / (experiment_id + ".tar.gz")

    assert path_exists(repository, experiment_storage_path)

    subprocess.run(
        ["replicate", "delete", "--force", experiment_id],
        cwd=tmpdir,
        env=env,
        check=True,
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
    assert not path_exists(repository, experiment_storage_path)
