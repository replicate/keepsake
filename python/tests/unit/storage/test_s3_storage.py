import pytest  # type: ignore
import boto3
import moto  # type: ignore
import os

from replicate.exceptions import DoesNotExistError
from replicate.storage.s3_storage import S3Storage

BUCKET_NAME = "mybucket"


@pytest.fixture
def mock_s3():
    mock = moto.mock_s3()
    mock.start()
    client = boto3.client("s3")
    client.create_bucket(Bucket=BUCKET_NAME)
    yield client
    mock.stop()


def mock_get(mock_s3, key):
    return mock_s3.get_object(Bucket=BUCKET_NAME, Key=key)["Body"].read()


def test_put_get(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME, root="")
    storage.put("foo/bar.txt", "nice")
    assert (
        mock_s3.get_object(Bucket=BUCKET_NAME, Key="foo/bar.txt")["Body"].read()
        == b"nice"
    )
    assert storage.get("foo/bar.txt") == b"nice"


def test_put_get_with_root(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME, root="someroot")
    storage.put("foo/bar.txt", "nice")
    assert mock_get(mock_s3, "someroot/foo/bar.txt") == b"nice"
    assert storage.get("foo/bar.txt") == b"nice"


def test_get_not_exists(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME, root="")
    with pytest.raises(DoesNotExistError):
        assert storage.get("foo/bar.txt")

    storage = S3Storage(bucket=BUCKET_NAME, root="someroot")
    with pytest.raises(DoesNotExistError):
        assert storage.get("foo/bar.txt")


def test_exists(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME, root="")
    assert not storage.exists("foo.txt")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="foo.txt")
    assert storage.exists("foo.txt")


def test_exists_with_root(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME, root="someroot")
    assert not storage.exists("foo.txt")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="someroot/foo.txt")
    assert storage.exists("foo.txt")


def test_delete_exists(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME, root="")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="foo.txt")
    storage.delete("foo.txt")
    assert not storage.exists("foo.txt")


def test_delete_not_exists(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME, root="")
    with pytest.raises(DoesNotExistError):
        storage.delete("foo.txt")


def test_delete(mock_s3):
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="foo.txt")
    storage = S3Storage(bucket=BUCKET_NAME, root="")
    assert storage.exists("foo.txt")
    storage.delete("foo.txt")
    assert not storage.exists("foo.txt")


def test_delete_with_root(mock_s3):
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="someroot/foo.txt")
    storage = S3Storage(bucket=BUCKET_NAME, root="someroot")
    assert storage.exists("foo.txt")
    storage.delete("foo.txt")
    assert not storage.exists("foo.txt")


def test_list(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME, root="")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="hello.txt")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="some/foo.txt")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="some/bar.txt")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="some/baz/qux.txt")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="some/baz/quux.txt")

    assert list(storage.list("")) == [
        {"name": "hello.txt", "type": "file"},
        {"name": "some", "type": "directory"},
    ]

    assert list(storage.list("some")) == [
        {"name": "bar.txt", "type": "file"},
        {"name": "baz", "type": "directory"},
        {"name": "foo.txt", "type": "file"},
    ]


def test_list_with_root(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME, root="someroot")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="someroot/hello.txt")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="someroot/some/foo.txt")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="someroot/some/bar.txt")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="someroot/some/baz/qux.txt")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="someroot/some/baz/quux.txt")

    assert list(storage.list("")) == [
        {"name": "hello.txt", "type": "file"},
        {"name": "some", "type": "directory"},
    ]

    assert list(storage.list("some")) == [
        {"name": "bar.txt", "type": "file"},
        {"name": "baz", "type": "directory"},
        {"name": "foo.txt", "type": "file"},
    ]


def test_put_path(mock_s3, tmpdir):
    storage = S3Storage(bucket=BUCKET_NAME, root="")

    for path in ["foo.txt", "bar/baz.txt", "qux.txt"]:
        abs_path = os.path.join(tmpdir, path)
        os.makedirs(os.path.dirname(abs_path), exist_ok=True)
        with open(abs_path, "w") as f:
            f.write("hello " + path)

    storage.put_path("folder", tmpdir)

    assert mock_get(mock_s3, "folder/foo.txt") == b"hello foo.txt"
    assert mock_get(mock_s3, "folder/qux.txt") == b"hello qux.txt"
    assert mock_get(mock_s3, "folder/bar/baz.txt") == b"hello bar/baz.txt"

    # single files
    storage.put_path("singlefile/foo.txt", os.path.join(tmpdir, "foo.txt"))
    assert mock_get(mock_s3, "singlefile/foo.txt") == b"hello foo.txt"


def test_put_path_with_root(mock_s3, tmpdir):
    storage = S3Storage(bucket=BUCKET_NAME, root="someroot")

    for path in ["foo.txt", "bar/baz.txt", "qux.txt"]:
        abs_path = os.path.join(tmpdir, path)
        os.makedirs(os.path.dirname(abs_path), exist_ok=True)
        with open(abs_path, "w") as f:
            f.write("hello " + path)

    storage.put_path("folder", tmpdir)

    assert mock_get(mock_s3, "someroot/folder/foo.txt") == b"hello foo.txt"
    assert mock_get(mock_s3, "someroot/folder/qux.txt") == b"hello qux.txt"
    assert mock_get(mock_s3, "someroot/folder/bar/baz.txt") == b"hello bar/baz.txt"

    # single files
    storage.put_path("singlefile/foo.txt", os.path.join(tmpdir, "foo.txt"))
    assert mock_get(mock_s3, "someroot/singlefile/foo.txt") == b"hello foo.txt"
