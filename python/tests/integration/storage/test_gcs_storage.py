import json
import os
import random
import string
import pytest  # type: ignore
from google.cloud import storage

import replicate
from replicate.exceptions import DoesNotExistError
from replicate.storage.gcs_storage import GCSStorage


@pytest.fixture(scope="function")
def temp_bucket():
    bucket_name = "replicate-test-" + "".join(
        random.choice(string.ascii_lowercase) for _ in range(20)
    )

    client = storage.Client()
    bucket = client.bucket(bucket_name)
    bucket.create()
    try:
        bucket.reload()
        assert bucket.exists()
        yield bucket_name
    finally:
        bucket.delete(force=True)


def test_put_get(temp_bucket):
    storage = GCSStorage(bucket=temp_bucket)
    storage.put("foo/bar.txt", "nice")
    assert storage.get("foo/bar.txt") == b"nice"


def test_get_not_exists(temp_bucket):
    storage = GCSStorage(bucket=temp_bucket)
    with pytest.raises(DoesNotExistError):
        assert storage.get("foo/bar.txt")


def test_put_directory(temp_bucket, tmpdir):
    storage = GCSStorage(bucket=temp_bucket)

    for path in ["foo.txt", "bar/baz.txt", "qux.txt"]:
        abs_path = os.path.join(tmpdir, path)
        os.makedirs(os.path.dirname(abs_path), exist_ok=True)
        with open(abs_path, "w") as f:
            f.write("hello " + path)

    storage.put_directory("folder", tmpdir)
    assert storage.get("folder/foo.txt") == b"hello foo.txt"
    assert storage.get("folder/qux.txt") == b"hello qux.txt"
    assert storage.get("folder/bar/baz.txt") == b"hello bar/baz.txt"
