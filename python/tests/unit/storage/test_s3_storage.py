import pytest  # type: ignore
import boto3
import moto  # type: ignore

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


def test_put_get(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME)
    storage.put("foo/bar.txt", "nice")
    assert storage.get("foo/bar.txt") == b"nice"


def test_get_not_exists(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME)
    with pytest.raises(DoesNotExistError):
        assert storage.get("foo/bar.txt")


def test_exists(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME)
    assert not storage.exists("foo.txt")
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="foo.txt")
    assert storage.exists("foo.txt")


def test_delete_exists(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME)
    mock_s3.put_object(Bucket=BUCKET_NAME, Key="foo.txt")
    storage.delete("foo.txt")
    assert not storage.exists("foo.txt")


def test_delete_not_exists(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME)
    with pytest.raises(DoesNotExistError):
        storage.delete("foo.txt")


def test_list(mock_s3):
    storage = S3Storage(bucket=BUCKET_NAME)
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
