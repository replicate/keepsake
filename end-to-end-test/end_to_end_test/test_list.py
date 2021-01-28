import json
import os
import subprocess
import pytest  # type: ignore
from dateutil.parser import parse as parse_date

from .utils import get_env


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
def test_list(
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

    with open(os.path.join(tmpdir, "keepsake.yaml"), "w") as f:
        f.write(
            """
repository: {repository}
""".format(
                repository=repository
            )
        )
    with open(os.path.join(tmpdir, "train.py"), "w") as f:
        f.write(
            """
import keepsake

def main():
    experiment = keepsake.init(path=".", params={"my_param": "my-value"})

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
            ["keepsake", "--verbose", "list", "--json"],
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
    assert parse_date(latest["created"]) > parse_date(exp["created"])
    assert latest["step"] == 2

    # test that --repository works
    experiments2 = json.loads(
        subprocess.run(
            ["keepsake", "--verbose", "ls", "--json", "--repository=" + repository],
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
from keepsake.project import Project
import sys

experiment_id = sys.argv[1]
experiment = Project().experiments.get(experiment_id)
experiment.checkpoint(path=".", step=3)
"""
        )
    subprocess.run(["python", "train2.py", exp["id"]], cwd=tmpdir, env=env, check=True)
    experiments = json.loads(
        subprocess.run(
            ["keepsake", "--verbose", "list", "--json"],
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
