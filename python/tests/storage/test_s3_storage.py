import json
import os
import pytest  # type: ignore
import boto3

import replicate
from replicate.hash import random_hash
from replicate.storage.s3_storage import S3Storage
from replicate.exceptions import DoesNotExistError

# Disable this test with -m "not external"
pytestmark = pytest.mark.external


@pytest.fixture(scope="function")
def temp_bucket():
    bucket_name = "replicate-test-" + random_hash()[:20]

    yield bucket_name

    s3 = boto3.resource("s3")
    bucket = s3.Bucket(bucket_name)
    bucket.objects.all().delete()
    bucket.delete()


def test_s3_experiment(temp_bucket, tmpdir):
    replicate_yaml_contents = "storage: s3://{bucket}".format(bucket=temp_bucket)

    with open(os.path.join(tmpdir, "replicate.yaml"), "w") as f:
        f.write(replicate_yaml_contents)

    current_workdir = os.getcwd()
    try:
        os.chdir(tmpdir)
        experiment = replicate.init(path=".", params={"foo": "bar"})
        checkpoint = experiment.checkpoint(
            path=".", step=10, metrics={"loss": 1.1, "baz": "qux"}
        )

        actual_experiment_meta = s3_read_json(
            temp_bucket,
            os.path.join("metadata", "experiments", experiment.id + ".json"),
        )
        # TODO(andreas): actually check values of host and user
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
            "config": {"python": "3.7", "storage": "s3://" + temp_bucket},
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
        }
        assert actual_experiment_meta == expected_experiment_meta

    finally:
        os.chdir(current_workdir)


def test_list(temp_bucket):
    storage = S3Storage(bucket=temp_bucket, root="")
    storage.put("foo", "nice")
    storage.put("some/bar", "nice")
    assert storage.list("") == ["foo"]
    assert storage.list("some") == ["some/bar"]


def s3_read(bucket, path):
    s3 = boto3.client("s3")
    return s3.get_object(Bucket=bucket, Key=path)["Body"].read()


def s3_read_json(bucket, path):
    return json.loads(s3_read(bucket, path))


def test_delete(temp_bucket, tmpdir):
    storage = S3Storage(bucket=temp_bucket, root="")

    storage.put("some/file", "nice")
    assert storage.get("some/file") == b"nice"

    storage.delete("some/file")
    with pytest.raises(DoesNotExistError):
        storage.get("some/file")


def test_delete_with_root(temp_bucket, tmpdir):
    storage = S3Storage(bucket=temp_bucket, root="my-root")

    storage.put("some/file", "nice")
    assert storage.get("some/file") == b"nice"

    storage.delete("some/file")
    with pytest.raises(DoesNotExistError):
        storage.get("some/file")
