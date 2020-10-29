import pytest  # type: ignore

from replicate.repository import repository_for_url
from replicate.repository.disk_repository import DiskRepository
from replicate.repository.s3_repository import S3Repository
from replicate.repository.gcs_repository import GCSRepository
from replicate.exceptions import UnknownRepositoryBackend


# parallel of go/pkg/repository/repository_test.go


def test_implicit_disk_repository():
    repository = repository_for_url("/foo/bar")
    assert isinstance(repository, DiskRepository)
    assert repository.root == "/foo/bar"

    repository = repository_for_url("foo/bar")
    assert isinstance(repository, DiskRepository)
    assert repository.root == "foo/bar"


def test_disk_repository():
    repository = repository_for_url("file:///foo/bar")
    assert isinstance(repository, DiskRepository)
    assert repository.root == "/foo/bar"

    repository = repository_for_url("file://foo/bar")
    assert isinstance(repository, DiskRepository)
    assert repository.root == "foo/bar"


def test_s3_repository():
    repository = repository_for_url("s3://my-bucket")
    assert isinstance(repository, S3Repository)
    assert repository.bucket_name == "my-bucket"
    assert repository.root == ""

    repository = repository_for_url("s3://my-bucket/foo")
    assert isinstance(repository, S3Repository)
    assert repository.bucket_name == "my-bucket"
    assert repository.root == "foo"


def test_gcs_repository():
    repository = repository_for_url("gs://my-bucket")
    assert isinstance(repository, GCSRepository)
    assert repository.bucket_name == "my-bucket"
    assert repository.root == ""

    repository = repository_for_url("gs://my-bucket/foo")
    assert isinstance(repository, GCSRepository)
    assert repository.bucket_name == "my-bucket"
    assert repository.root == "foo"


def test_unknown_repository():
    with pytest.raises(UnknownRepositoryBackend):
        repository_for_url("foo://my-bucket")
