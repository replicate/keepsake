from glob import glob
import random
import json
import os
import subprocess
import pytest  # type: ignore


@pytest.mark.parametrize(
    "storage_backend", ["gcs", "s3", "file"],
)
def test_checkout(storage_backend, tmpdir, temp_bucket, tmpdir_factory):
    tmpdir = str(tmpdir)
    if storage_backend == "s3":
        storage = "s3://" + temp_bucket
    if storage_backend == "gcs":
        storage = "gs://" + temp_bucket
    elif storage_backend == "file":
        storage = "file://" + str(tmpdir_factory.mktemp("storage"))

    rand = str(random.randint(0, 100000))
    os.mkdir(os.path.join(tmpdir, rand))
    with open(os.path.join(tmpdir, rand, rand), "w") as f:
        f.write(rand)

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
    experiment = replicate.init()
    experiment.commit()

if __name__ == "__main__":
    main()
"""
        )

    env = os.environ
    env["PATH"] = "/usr/local/bin:" + os.environ["PATH"]

    cmd = ["python", "train.py"]
    return_code = subprocess.Popen(cmd, cwd=tmpdir, env=env).wait()
    assert return_code == 0

    experiments = json.loads(
        subprocess.check_output(
            ["replicate", "list", "--json"], cwd=tmpdir, env=env,
        )
    )
    assert len(experiments) == 1

    exp = experiments[0]
    latest_id = exp["latest_commit"]["id"]

    output_dir = str(tmpdir_factory.mktemp("output"))
    subprocess.check_output(
        ["replicate", "checkout", "-o", output_dir, latest_id], cwd=tmpdir, env=env,
    )

    with open(os.path.join(output_dir, rand, rand)) as f:
        assert f.read() == rand

    actual_paths = [
        os.path.relpath(path, output_dir) for path in glob(output_dir + "/*")
    ]
    expected_paths = ["replicate.yaml", "train.py", rand]
    assert set(actual_paths) == set(expected_paths)
