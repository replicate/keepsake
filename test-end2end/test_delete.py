import json
import os
import subprocess
import pytest  # type: ignore


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
def test_list(storage_backend, use_root, tmpdir, temp_bucket, tmpdir_factory):
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
    experiment = replicate.init(params={"my_param": "my-value"})

    for step in range(3):
        experiment.checkpoint(path=".", step=step)

if __name__ == "__main__":
    main()
"""
        )

    env = os.environ
    env["PATH"] = "/usr/local/bin:" + os.environ["PATH"]

    cmd = ["python", "train.py", "--foo"]
    subprocess.run(cmd, cwd=tmpdir, env=env)

    experiments = json.loads(
        subprocess.run(
            ["replicate", "list", "--json"], cwd=tmpdir, env=env, capture_output=True
        ).stdout
    )
    assert len(experiments) == 1
    assert experiments[0]["num_checkpoints"] == 3

    checkpoint_id = experiments[0]["latest_checkpoint"]["id"]
    subprocess.run(["replicate", "delete", checkpoint_id], cwd=tmpdir, env=env)

    experiments = json.loads(
        subprocess.run(
            ["replicate", "list", "--json"], cwd=tmpdir, env=env, capture_output=True
        ).stdout
    )
    assert len(experiments) == 1
    assert experiments[0]["num_checkpoints"] == 2

    subprocess.run(["replicate", "delete", experiments[0]["id"]], cwd=tmpdir, env=env)

    experiments = json.loads(
        subprocess.run(
            ["replicate", "list", "--json"], cwd=tmpdir, env=env, capture_output=True
        ).stdout
    )
    assert len(experiments) == 0
