import urllib.request
from glob import glob
import random
import json
import os
import subprocess
import pytest  # type: ignore
import shlex
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
import os
import keepsake

def main():
    experiment = keepsake.init(path=".")
    os.mkdir("data")
    with open("data/weights", "w") as fh:
        fh.write("42 lbs")
    experiment.checkpoint(path="data/")

if __name__ == "__main__":
    main()
"""
        )

    env = get_env()
    cmd = ["python", "train.py"]
    subprocess.run(cmd, cwd=tmpdir, env=env, check=True)

    experiments = json.loads(
        subprocess.run(
            ["keepsake", "ls", "--json"],
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
        f"keepsake checkout -o {shlex.quote(output_dir)} {exp['id']}",
        cwd=tmpdir,
        env=env,
        check=True,
        shell=True,
    )

    with open(os.path.join(output_dir, rand, rand)) as f:
        assert f.read() == rand

    # Checkout out experiment checks out latest checkpoint
    with open(os.path.join(output_dir, "data/weights")) as f:
        assert f.read() == "42 lbs"

    actual_paths = [
        os.path.relpath(path, output_dir) for path in glob(output_dir + "/*")
    ]
    expected_paths = ["keepsake.yaml", "train.py", "data", rand, "cicada.ogg"]
    assert set(actual_paths) == set(expected_paths)

    # checking out checkpoint
    latest_id = exp["latest_checkpoint"]["id"]

    output_dir = str(tmpdir_factory.mktemp("output"))
    subprocess.run(
        ["keepsake", "checkout", "-o", output_dir, latest_id],
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
        "keepsake.yaml",
        "train.py",
        "data",
        rand,
        "cicada.ogg",
    ]
    assert set(actual_paths) == set(expected_paths)


def test_checkout_no_experiment_path(tmpdir, temp_bucket_factory, tmpdir_factory):
    tmpdir = str(tmpdir)
    repository = "file://" + str(tmpdir_factory.mktemp("repository"))

    rand = str(random.randint(0, 100000))
    os.mkdir(os.path.join(tmpdir, rand))
    with open(os.path.join(tmpdir, rand, rand), "w") as f:
        f.write(rand)

    with open(os.path.join(tmpdir, "foo.txt"), "w") as f:
        f.write("foo bar")

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
import os
import keepsake

def main():
    experiment = keepsake.init()
    experiment.checkpoint(path="foo.txt")

if __name__ == "__main__":
    main()
"""
        )

    env = get_env()
    cmd = ["python", "train.py"]
    subprocess.run(cmd, cwd=tmpdir, env=env, check=True)

    experiments = json.loads(
        subprocess.run(
            ["keepsake", "ls", "--json"],
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
        ["keepsake", "checkout", "-o", output_dir, exp["id"]],
        cwd=tmpdir,
        env=env,
        check=True,
    )

    # Checkout out experiment checks out latest checkpoint
    with open(os.path.join(output_dir, "foo.txt")) as f:
        assert f.read() == "foo bar"

    actual_paths = [
        os.path.relpath(path, output_dir) for path in glob(output_dir + "/*")
    ]
    expected_paths = ["foo.txt"]
    assert set(actual_paths) == set(expected_paths)

    # checking out checkpoint
    latest_id = exp["latest_checkpoint"]["id"]

    output_dir = str(tmpdir_factory.mktemp("output"))
    subprocess.run(
        ["keepsake", "checkout", "-o", output_dir, latest_id],
        cwd=tmpdir,
        env=env,
        check=True,
    )

    with open(os.path.join(output_dir, "foo.txt")) as f:
        assert f.read() == "foo bar"

    actual_paths = [
        os.path.relpath(path, output_dir) for path in glob(output_dir + "/*")
    ]
    expected_paths = ["foo.txt"]
    assert set(actual_paths) == set(expected_paths)


def test_checkout_when_files_exist(tmpdir, tmpdir_factory):
    tmpdir = str(tmpdir)
    repository = "file://" + str(tmpdir_factory.mktemp("repository"))

    rand = str(random.randint(0, 100000))
    os.mkdir(os.path.join(tmpdir, rand))
    with open(os.path.join(tmpdir, rand, rand), "w") as f:
        f.write(rand)

    with open(os.path.join(tmpdir, "foo.txt"), "w") as f:
        f.write("original")

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
import os
import keepsake

def main():
    experiment = keepsake.init()
    experiment.checkpoint(path="foo.txt")

if __name__ == "__main__":
    main()
"""
        )

    env = get_env()
    cmd = ["python", "train.py"]
    subprocess.run(cmd, cwd=tmpdir, env=env, check=True)

    experiments = json.loads(
        subprocess.run(
            ["keepsake", "ls", "--json"],
            cwd=tmpdir,
            env=env,
            capture_output=True,
            check=True,
        ).stdout
    )
    assert len(experiments) == 1

    exp = experiments[0]

    with open(os.path.join(tmpdir, "foo.txt"), "w") as f:
        f.write("new")

    # stdin is closed
    result = subprocess.run(
        f"keepsake checkout {exp['id']}", cwd=tmpdir, env=env, check=False, shell=True,
    )
    assert result.returncode > 0

    # Checkout does not overwrite
    with open(os.path.join(tmpdir, "foo.txt")) as f:
        assert f.read() == "new"

    # don't overwrite
    result = subprocess.run(
        f"keepsake checkout {exp['id']}",
        cwd=tmpdir,
        env=env,
        shell=True,
        check=False,
        input=b"n\n",
    )
    assert result.returncode > 0

    # Checkout does not overwrite
    with open(os.path.join(tmpdir, "foo.txt")) as f:
        assert f.read() == "new"

    # do overwrite
    subprocess.run(
        f"keepsake checkout {exp['id']}",
        cwd=tmpdir,
        env=env,
        check=True,
        shell=True,
        input=b"y\n",
    )

    # Checkout does overwrite!
    with open(os.path.join(tmpdir, "foo.txt")) as f:
        assert f.read() == "original"
