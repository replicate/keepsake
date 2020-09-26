import os
import pathlib
import pytest  # type: ignore
import tempfile

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
        actual = list(storage.list(""))
        actual.sort(key=lambda x: x["name"])
        expected = [
            {"name": "foo", "type": "file"},
            {"name": "some", "type": "directory"},
        ]
        assert actual == expected
        assert list(storage.list("some")) == [{"name": "bar", "type": "file"}]


def test_delete():
    with tempfile.TemporaryDirectory() as tmpdir:
        storage = DiskStorage(root=tmpdir)
        storage.put("some/file", "nice")
        assert storage.get("some/file") == b"nice"
        storage.delete("some/file")
        with pytest.raises(DoesNotExistError):
            storage.get("some/file")


def test_put_path():
    with tempfile.TemporaryDirectory() as src:
        src_path = pathlib.Path(src)
        for path in ["foo.txt", "bar/baz.txt", "qux.txt"]:
            abs_path = src_path / path
            abs_path.parent.mkdir(parents=True, exist_ok=True)
            with open(abs_path, "w") as f:
                f.write("hello " + path)

        with tempfile.TemporaryDirectory() as root:
            root_path = pathlib.Path(root)
            storage = DiskStorage(root=root)
            storage.put_path("somedir", src)
            assert open(root_path / "somedir/foo.txt").read() == "hello foo.txt"
            assert open(root_path / "somedir/qux.txt").read() == "hello qux.txt"
            assert open(root_path / "somedir/bar/baz.txt").read() == "hello bar/baz.txt"

            # single files
            storage.put_path("singlefile/foo.txt", os.path.join(src, "foo.txt"))
            assert open(root_path / "singlefile/foo.txt").read() == "hello foo.txt"
