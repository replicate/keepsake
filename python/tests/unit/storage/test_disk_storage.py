import os
import pathlib
import pytest  # type: ignore
import tarfile
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
        assert storage.list("") == ["foo"]
        assert storage.list("some") == ["some/bar"]


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
            storage.put_path(src, "somedir")
            assert open(root_path / "somedir/foo.txt").read() == "hello foo.txt"
            assert open(root_path / "somedir/qux.txt").read() == "hello qux.txt"
            assert open(root_path / "somedir/bar/baz.txt").read() == "hello bar/baz.txt"

            # single files
            storage.put_path(os.path.join(src, "foo.txt"), "singlefile/foo.txt")
            assert open(root_path / "singlefile/foo.txt").read() == "hello foo.txt"


def test_put_path_tar():
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
            storage.put_path_tar(src, "dest.tar.gz", "")

            with tempfile.TemporaryDirectory() as out:
                out = pathlib.Path(out)
                with tarfile.open(root_path / "dest.tar.gz") as tar:
                    tar.extractall(out)
                assert open(out / "dest/foo.txt").read() == "hello foo.txt"
