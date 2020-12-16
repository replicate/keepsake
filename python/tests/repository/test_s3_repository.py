import json
import os
import tempfile
from pathlib import Path
import pytest  # type: ignore
import boto3  # type: ignore

import replicate
from replicate.hash import random_hash
from replicate.repository.s3_repository import S3Repository
from replicate.exceptions import DoesNotExistError

# Disable this test with -m "not external"
pytestmark = pytest.mark.external


@pytest.fixture(scope="function")
def temp_bucket():
    bucket_name = "replicate-test-unit-" + random_hash()[:20]

    yield bucket_name

    s3 = boto3.resource("s3")
    bucket = s3.Bucket(bucket_name)
    bucket.objects.all().delete()
    bucket.delete()


def test_s3_experiment(temp_bucket, tmpdir):
    replicate_yaml_contents = "repository: s3://{bucket}".format(bucket=temp_bucket)

    with open(os.path.join(tmpdir, "replicate.yaml"), "w") as f:
        f.write(replicate_yaml_contents)

    current_workdir = os.getcwd()
    try:
        os.chdir(tmpdir)
        experiment = replicate.init(
            path=".", params={"foo": "bar"}, disable_heartbeat=True
        )
        checkpoint = experiment.checkpoint(
            path=".", step=10, metrics={"loss": 1.1, "baz": "qux"}
        )

        actual_experiment_meta = s3_read_json(
            temp_bucket,
            os.path.join("metadata", "experiments", experiment.id + ".json"),
        )
        assert "host" in actual_experiment_meta
        assert "user" in actual_experiment_meta
        del actual_experiment_meta["host"]
        del actual_experiment_meta["user"]
        del actual_experiment_meta["command"]
        del actual_experiment_meta["python_packages"]

        expected_experiment_meta = {
            "id": experiment.id,
            "created": experiment.created.isoformat() + "Z",
            "params": {"foo": "bar"},
            "config": {"repository": "s3://" + temp_bucket},
            "path": ".",
            "checkpoints": [
                {
                    "id": checkpoint.id,
                    "created": checkpoint.created.isoformat() + "Z",
                    "step": 10,
                    "metrics": {"loss": 1.1, "baz": "qux"},
                    "path": ".",
                    "primary_metric": None,
                }
            ],
            "replicate_version": replicate.__version__,
        }
        assert actual_experiment_meta == expected_experiment_meta

    finally:
        os.chdir(current_workdir)


def test_list(temp_bucket):
    repository = S3Repository(bucket=temp_bucket, root="")
    repository.put("foo", "nice")
    repository.put("some/bar", "nice")
    assert repository.list("") == ["foo"]
    assert repository.list("some") == ["some/bar"]


def s3_read(bucket, path):
    s3 = boto3.client("s3")
    return s3.get_object(Bucket=bucket, Key=path)["Body"].read()


def s3_read_json(bucket, path):
    return json.loads(s3_read(bucket, path))


def test_delete(temp_bucket, tmpdir):
    repository = S3Repository(bucket=temp_bucket, root="")

    repository.put("some/file", "nice")
    assert repository.get("some/file") == b"nice"

    repository.delete("some/file")
    with pytest.raises(DoesNotExistError):
        repository.get("some/file")


def test_delete_with_root(temp_bucket, tmpdir):
    repository = S3Repository(bucket=temp_bucket, root="my-root")

    repository.put("some/file", "nice")
    assert repository.get("some/file") == b"nice"

    repository.delete("some/file")
    with pytest.raises(DoesNotExistError):
        repository.get("some/file")


def test_get_put_path_tar(temp_bucket):
    with tempfile.TemporaryDirectory() as src:
        src_path = Path(src)
        for path in ["foo.txt", "bar/baz.txt", "qux.txt"]:
            abs_path = src_path / path
            abs_path.parent.mkdir(parents=True, exist_ok=True)
            with open(abs_path, "w") as f:
                f.write("hello " + path)

        repository = S3Repository(bucket=temp_bucket, root="")
        repository.put_path_tar(src, "dest.tar.gz", "")

    with tempfile.TemporaryDirectory() as out:
        repository.get_path_tar("dest.tar.gz", out)
        out = Path(out)
        assert open(out / "foo.txt").read() == "hello foo.txt"
