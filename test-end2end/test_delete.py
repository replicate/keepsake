import json
import os
import subprocess
import pytest  # type: ignore


@pytest.mark.parametrize(
    "storage_backend", ["gcs", "s3", pytest.param("file", marks=pytest.mark.fast)],
)
def test_list(storage_backend, tmpdir, temp_bucket, tmpdir_factory):
    tmpdir = str(tmpdir)
    if storage_backend == "s3":
        storage = "s3://" + temp_bucket
    if storage_backend == "gcs":
        storage = "gs://" + temp_bucket
    elif storage_backend == "file":
        storage = "file://" + str(tmpdir_factory.mktemp("storage"))

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
        experiment.commit(path=".", step=step)

if __name__ == "__main__":
    main()
"""
        )
    # TODO: upload Python library
    with open(os.path.join(tmpdir, "requirements.txt"), "w") as f:
        f.write("replicate==0.1.8")

    env = os.environ
    env["PATH"] = "/usr/local/bin:" + os.environ["PATH"]

    cmd = ["python", "train.py", "--foo"]
    return_code = subprocess.Popen(cmd, cwd=tmpdir, env=env).wait()
    assert return_code == 0

    experiments = json.loads(
        subprocess.check_output(["replicate", "list", "--json"], cwd=tmpdir, env=env)
    )
    assert len(experiments) == 1
    assert experiments[0]["num_commits"] == 3

    commit_id = experiments[0]["latest_commit"]["id"]
    subprocess.run(["replicate", "delete", commit_id], cwd=tmpdir, env=env)

    experiments = json.loads(
        subprocess.check_output(["replicate", "list", "--json"], cwd=tmpdir, env=env)
    )
    assert len(experiments) == 1
    assert experiments[0]["num_commits"] == 2

    subprocess.run(["replicate", "delete", experiments[0]["id"]], cwd=tmpdir, env=env)

    experiments = json.loads(
        subprocess.check_output(["replicate", "list", "--json"], cwd=tmpdir, env=env)
    )
    assert len(experiments) == 0
