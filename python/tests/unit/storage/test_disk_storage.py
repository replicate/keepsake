import tempfile
import pytest  # type: ignore

from replicate.exceptions import DoesNotExistError
from replicate.storage.disk_storage import DiskStorage


def test_put_get():
    with tempfile.TemporaryDirectory() as tmpdir:
        storage = DiskStorage(root=tmpdir)
        storage.put("some/file", "nice")
        assert storage.get("some/file") == b"nice"
        with pytest.raises(DoesNotExistError):
            storage.get("not/here")


def test_list():
    with tempfile.TemporaryDirectory() as tmpdir:
        storage = DiskStorage(root=tmpdir)
        storage.put("foo", "nice")
        storage.put("some/bar", "nice")
        assert list(storage.list("")) == [
            {"name": "foo", "type": "file"},
            {"name": "some", "type": "directory"},
        ]
        assert list(storage.list("some")) == [{"name": "bar", "type": "file"}]


def test_delete():
    with tempfile.TemporaryDirectory() as tmpdir:
        storage = DiskStorage(root=tmpdir)
        storage.put("some/file", "nice")
        assert storage.get("some/file") == b"nice"
        storage.delete("some/file")
        with pytest.raises(DoesNotExistError):
            storage.get("some/file")
