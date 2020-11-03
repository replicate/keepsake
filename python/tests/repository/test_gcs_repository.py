import os
import random
import string
import tempfile
from pathlib import Path
import pytest  # type: ignore
from google.cloud import storage
from google.api_core.exceptions import NotFound

from replicate.exceptions import DoesNotExistError
from replicate.repository.gcs_repository import GCSRepository


# Disable this test with -m "not external"
pytestmark = pytest.mark.external

# We only create one bucket for these tests, because Google Cloud rate limits creating buckets
# https://cloud.google.com/storage/quotas
# This means these tests can't be run in parallel
@pytest.fixture(scope="session")
def temp_bucket_create():
    bucket_name = "replicate-test-" + "".join(
        random.choice(string.ascii_lowercase) for _ in range(20)
    )

    client = storage.Client()
    bucket = client.create_bucket(bucket_name)
    try:
        bucket.reload()
        assert bucket.exists()
        yield bucket
    finally:
        bucket.delete(force=True)


@pytest.fixture(scope="function")
def temp_bucket(temp_bucket_create):
    bucket = temp_bucket_create
    # Clear bucket before each test
    blobs = bucket.list_blobs()
    for blob in blobs:
        blob.delete()
    yield bucket


def test_put_get(temp_bucket):
    repository = GCSRepository(bucket=temp_bucket.name, root="")
    repository.put("foo/bar.txt", "nice")
    assert temp_bucket.blob("foo/bar.txt").download_as_bytes() == b"nice"
    assert repository.get("foo/bar.txt") == b"nice"


def test_put_get_with_root(temp_bucket):
    repository = GCSRepository(bucket=temp_bucket.name, root="someroot")
    repository.put("foo/bar.txt", "nice")
    assert temp_bucket.blob("someroot/foo/bar.txt").download_as_bytes() == b"nice"
    assert repository.get("foo/bar.txt") == b"nice"


def test_get_not_exists(temp_bucket):
    repository = GCSRepository(bucket=temp_bucket.name, root="")
    with pytest.raises(DoesNotExistError):
        assert repository.get("foo/bar.txt")


def test_list(temp_bucket):
    repository = GCSRepository(bucket=temp_bucket.name, root="")
    repository.put("foo", "nice")
    repository.put("some/bar", "nice")
    assert repository.list("") == ["foo"]
    assert repository.list("some") == ["some/bar"]


def test_put_path(temp_bucket, tmpdir):
    repository = GCSRepository(bucket=temp_bucket.name, root="")

    for path in ["foo.txt", "bar/baz.txt", "qux.txt"]:
        abs_path = os.path.join(tmpdir, path)
        os.makedirs(os.path.dirname(abs_path), exist_ok=True)
        with open(abs_path, "w") as f:
            f.write("hello " + path)

    repository.put_path(tmpdir, "folder")
    assert temp_bucket.blob("folder/foo.txt").download_as_bytes() == b"hello foo.txt"
    assert temp_bucket.blob("folder/qux.txt").download_as_bytes() == b"hello qux.txt"
    assert (
        temp_bucket.blob("folder/bar/baz.txt").download_as_bytes()
        == b"hello bar/baz.txt"
    )

    # single files
    repository.put_path(os.path.join(tmpdir, "foo.txt"), "singlefile/foo.txt")
    assert (
        temp_bucket.blob("singlefile/foo.txt").download_as_bytes() == b"hello foo.txt"
    )


def test_put_path_with_root(temp_bucket, tmpdir):
    repository = GCSRepository(bucket=temp_bucket.name, root="someroot")

    for path in ["foo.txt", "bar/baz.txt", "qux.txt"]:
        abs_path = os.path.join(tmpdir, path)
        os.makedirs(os.path.dirname(abs_path), exist_ok=True)
        with open(abs_path, "w") as f:
            f.write("hello " + path)

    repository.put_path(tmpdir, "folder")
    assert (
        temp_bucket.blob("someroot/folder/foo.txt").download_as_bytes()
        == b"hello foo.txt"
    )
    assert (
        temp_bucket.blob("someroot/folder/qux.txt").download_as_bytes()
        == b"hello qux.txt"
    )
    assert (
        temp_bucket.blob("someroot/folder/bar/baz.txt").download_as_bytes()
        == b"hello bar/baz.txt"
    )

    # single files
    repository.put_path(os.path.join(tmpdir, "foo.txt"), "singlefile/foo.txt")
    assert (
        temp_bucket.blob("someroot/singlefile/foo.txt").download_as_bytes()
        == b"hello foo.txt"
    )


def test_replicateignore(temp_bucket, tmpdir):
    repository = GCSRepository(bucket=temp_bucket.name, root="")

    for path in [
        "foo.txt",
        "bar/baz.txt",
        "bar/quux.xyz",
        "bar/new-qux.txt",
        "qux.xyz",
    ]:
        abs_path = os.path.join(tmpdir, path)
        os.makedirs(os.path.dirname(abs_path), exist_ok=True)
        with open(abs_path, "w") as f:
            f.write("hello " + path)

    with open(os.path.join(tmpdir, ".replicateignore"), "w") as f:
        f.write(
            """
# this is a comment
baz.txt
*.xyz
"""
        )

    repository.put_path(tmpdir, "folder")
    assert temp_bucket.blob("folder/foo.txt").download_as_bytes() == b"hello foo.txt"
    assert (
        temp_bucket.blob("folder/bar/new-qux.txt").download_as_bytes()
        == b"hello bar/new-qux.txt"
    )
    with pytest.raises(NotFound):
        temp_bucket.blob("folder/bar/baz.txt").download_as_bytes()
    with pytest.raises(NotFound):
        temp_bucket.blob("folder/qux.xyz").download_as_bytes()
    with pytest.raises(NotFound):
        temp_bucket.blob("folder/bar/quux.xyz").download_as_bytes()


def test_delete(temp_bucket, tmpdir):
    repository = GCSRepository(bucket=temp_bucket.name, root="")

    repository.put("some/file", "nice")
    assert repository.get("some/file") == b"nice"

    repository.delete("some/file")
    with pytest.raises(DoesNotExistError):
        repository.get("some/file")


def test_delete_with_root(temp_bucket, tmpdir):
    repository = GCSRepository(bucket=temp_bucket.name, root="my-root")

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

        repository = GCSRepository(bucket=temp_bucket.name, root="")
        repository.put_path_tar(src, "dest.tar.gz", "")

    with tempfile.TemporaryDirectory() as out:
        repository.get_path_tar("dest.tar.gz", out)
        out = Path(out)
        assert open(out / "foo.txt").read() == "hello foo.txt"
