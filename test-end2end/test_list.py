import json
import os
import subprocess
import pytest  # type: ignore


@pytest.mark.parametrize(
    "storage_backend,use_replicate_run",
    [
        ("gcs", False),
        ("s3", False),
        ("file", False),
        ("file", True),
        ("undefined", False),
    ],
)
def test_list(storage_backend, use_replicate_run, tmpdir, temp_bucket, tmpdir_factory):
    tmpdir = str(tmpdir)
    if storage_backend == "s3":
        storage = "s3://" + temp_bucket
    if storage_backend == "gcs":
        storage = "gs://" + temp_bucket
    elif storage_backend == "file":
        storage = "file://" + str(tmpdir_factory.mktemp("storage"))
    elif storage_backend == "undefined":
        storage = str(tmpdir_factory.mktemp("storage"))

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
    experiment = replicate.init(my_param="my-value")

    for step in range(3):
        experiment.commit(step=step)

if __name__ == "__main__":
    main()
"""
        )
    # TODO: upload Python library
    with open(os.path.join(tmpdir, "requirements.txt"), "w") as f:
        f.write("replicate==0.1.8")

    env = os.environ
    env["PATH"] = "/usr/local/bin:" + os.environ["PATH"]

    if use_replicate_run:
        cmd = ["replicate", "run", "train.py", "--foo"]
    else:
        cmd = ["python", "train.py", "--foo"]
    return_code = subprocess.Popen(cmd, cwd=tmpdir, env=env).wait()
    assert return_code == 0

    experiments = json.loads(
        subprocess.check_output(["replicate", "list", "--json"], cwd=tmpdir, env=env,)
    )
    assert len(experiments) == 1

    exp = experiments[0]
    latest = exp["latest_commit"]
    assert len(exp["id"]) == 64
    assert exp["params"] == {"my_param": "my-value"}
    assert exp["num_commits"] == 3
    if use_replicate_run:
        assert exp["command"] == "python -u train.py --foo"
    else:
        assert exp["command"] == "train.py --foo"
    assert len(latest["id"]) == 64
    # FIXME: now rfc3339 strings
    # assert latest["created"] > exp["created"]
    assert latest["step"] == 2

    # test that --storage-url works
    experiments2 = json.loads(
        subprocess.check_output(
            ["replicate", "list", "--json", "--storage-url=" + storage],
            cwd=tmpdir_factory.mktemp("list"),
            env=env,
        )
    )
    assert experiments2 == experiments
