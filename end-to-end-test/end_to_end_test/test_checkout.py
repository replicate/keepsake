import urllib.request
from glob import glob
import random
import json
import os
import subprocess
import pytest  # type: ignore

from .utils import get_env, PYTHON_PATH


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
def test_checkout(
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

    rand = str(random.randint(0, 100000))
    os.mkdir(os.path.join(tmpdir, rand))
    with open(os.path.join(tmpdir, rand, rand), "w") as f:
        f.write(rand)

    # big file (7.1MB)
    cicada_url = "https://storage.googleapis.com/replicate-public/cicada.ogg"
    urllib.request.urlretrieve(cicada_url, os.path.join(tmpdir, "cicada.ogg"))

    with open(os.path.join(tmpdir, "replicate.yaml"), "w") as f:
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
import os
import replicate

def main():
    experiment = replicate.init(path=".")
    os.mkdir("data")
    with open("data/weights", "w") as fh:
        fh.write("42 lbs")
    experiment.checkpoint(path="data/")

if __name__ == "__main__":
    main()
"""
        )

    env = get_env()
    cmd = [PYTHON_PATH, "train.py"]
    subprocess.run(cmd, cwd=tmpdir, env=env, check=True)

    experiments = json.loads(
        subprocess.run(
            ["replicate", "ls", "--json"],
            cwd=tmpdir,
            env=env,
            capture_output=True,
            check=True,
        ).stdout
    )
    assert len(experiments) == 1

    exp = experiments[0]

    # checking out experiment
    output_dir = str(tmpdir_factory.mktemp("output"))
    subprocess.run(
        ["replicate", "checkout", "-o", output_dir, exp["id"]],
        cwd=tmpdir,
        env=env,
        check=True,
    )

    with open(os.path.join(output_dir, rand, rand)) as f:
        assert f.read() == rand

    # Checkout out experiment checks out latest checkpoint
    with open(os.path.join(output_dir, "data/weights")) as f:
        assert f.read() == "42 lbs"

    actual_paths = [
        os.path.relpath(path, output_dir) for path in glob(output_dir + "/*")
    ]
    expected_paths = ["replicate.yaml", "train.py", "data", rand, "cicada.ogg"]
    assert set(actual_paths) == set(expected_paths)

    # checking out checkpoint
    latest_id = exp["latest_checkpoint"]["id"]

    output_dir = str(tmpdir_factory.mktemp("output"))
    subprocess.run(
        ["replicate", "checkout", "-o", output_dir, latest_id],
        cwd=tmpdir,
        env=env,
        check=True,
    )

    with open(os.path.join(output_dir, rand, rand)) as f:
        assert f.read() == rand

    with open(os.path.join(output_dir, "data/weights")) as f:
        assert f.read() == "42 lbs"

    actual_paths = [
        os.path.relpath(path, output_dir) for path in glob(output_dir + "/*")
    ]
    expected_paths = [
        "replicate.yaml",
        "train.py",
        "data",
        rand,
        "cicada.ogg",
    ]
    assert set(actual_paths) == set(expected_paths)
