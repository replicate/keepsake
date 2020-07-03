import pytest  # type: ignore

from replicate.storage import storage_for_url
from replicate.storage.disk_storage import DiskStorage
from replicate.storage.s3_storage import S3Storage
from replicate.exceptions import UnknownStorageBackend


def test_implicit_disk_storage():
    storage = storage_for_url("/foo/bar")
    assert isinstance(storage, DiskStorage)
    assert storage.root == "/foo/bar"


def test_disk_storage():
    storage = storage_for_url("file:///foo/bar")
    assert isinstance(storage, DiskStorage)
    assert storage.root == "/foo/bar"


def test_s3_storage():
    storage = storage_for_url("s3://my-bucket")
    assert isinstance(storage, S3Storage)
    assert storage.bucket == "my-bucket"


def test_unknown_storage():
    with pytest.raises(UnknownStorageBackend):
        storage_for_url("foo://my-bucket")
