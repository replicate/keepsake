import os
import pathlib
import pytest  # type: ignore
import tarfile
import tempfile

from replicate.exceptions import DoesNotExistError
from replicate.repository.disk_repository import DiskRepository


def test_put_get():
    with tempfile.TemporaryDirectory() as tmpdir:
        repository = DiskRepository(root=tmpdir)
        repository.put("some/file", "nice")
        assert repository.get("some/file") == b"nice"
        with pytest.raises(DoesNotExistError):
            repository.get("not/here")


def test_list():
    with tempfile.TemporaryDirectory() as tmpdir:
        repository = DiskRepository(root=tmpdir)
        repository.put("foo", "nice")
        repository.put("some/bar", "nice")
        assert repository.list("") == ["foo"]
        assert repository.list("some") == ["some/bar"]


def test_delete():
    with tempfile.TemporaryDirectory() as tmpdir:
        repository = DiskRepository(root=tmpdir)
        repository.put("some/file", "nice")
        assert repository.get("some/file") == b"nice"
        repository.delete("some/file")
        with pytest.raises(DoesNotExistError):
            repository.get("some/file")


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
            repository = DiskRepository(root=root)
            repository.put_path(src, "somedir")
            assert open(root_path / "somedir/foo.txt").read() == "hello foo.txt"
            assert open(root_path / "somedir/qux.txt").read() == "hello qux.txt"
            assert open(root_path / "somedir/bar/baz.txt").read() == "hello bar/baz.txt"

            # single files
            repository.put_path(os.path.join(src, "foo.txt"), "singlefile/foo.txt")
            assert open(root_path / "singlefile/foo.txt").read() == "hello foo.txt"


def test_get_put_path_tar():
    with tempfile.TemporaryDirectory() as src:
        src_path = pathlib.Path(src)
        for path in ["foo.txt", "bar/baz.txt", "qux.txt"]:
            abs_path = src_path / path
            abs_path.parent.mkdir(parents=True, exist_ok=True)
            with open(abs_path, "w") as f:
                f.write("hello " + path)

        with tempfile.TemporaryDirectory() as root:
            root_path = pathlib.Path(root)
            repository = DiskRepository(root=root)
            repository.put_path_tar(src, "dest.tar.gz", "")

            with tempfile.TemporaryDirectory() as out:
                out = pathlib.Path(out)
                with tarfile.open(root_path / "dest.tar.gz") as tar:
                    tar.extractall(out)
                assert open(out / "dest/foo.txt").read() == "hello foo.txt"

            with tempfile.TemporaryDirectory() as out:
                repository.get_path_tar("dest.tar.gz", out)
                out = pathlib.Path(out)
                assert open(out / "foo.txt").read() == "hello foo.txt"
