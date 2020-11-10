import os
import tempfile

import pytest

from replicate.project import get_project_dir, Project, ProjectSpec
from replicate.exceptions import ConfigNotFoundError, CorruptedProjectSpec


@pytest.fixture
def temp_workdir_in_subdir():
    orig_cwd = os.getcwd()
    try:
        with tempfile.TemporaryDirectory() as tmpdir:
            workdir = os.path.join(tmpdir, "foo", "bar")
            os.makedirs(workdir)
            os.chdir(workdir)
            yield
    finally:
        os.chdir(orig_cwd)


def test_get_project_dir(temp_workdir_in_subdir):
    # use getcwd instead of tempdir from fixture, because on OS X getcwd doesn't return same thing passed to chdir
    root = os.path.abspath(os.path.join(os.getcwd(), "../../"))

    # replicate.yaml in current directory
    open(os.path.join(root, "foo/bar/replicate.yaml"), "w").write("")
    assert get_project_dir() == os.path.join(root, "foo/bar")
    os.unlink(os.path.join(root, "foo/bar/replicate.yaml"))

    # up a directory
    open(os.path.join(root, "foo/replicate.yaml"), "w").write("")
    assert get_project_dir() == os.path.join(root, "foo")
    os.unlink(os.path.join(root, "foo/replicate.yaml"))

    # up two directories
    open(os.path.join(root, "replicate.yaml"), "w").write("")
    assert get_project_dir() == root
    os.unlink(os.path.join(root, "replicate.yaml"))

    # missing replicate.yaml
    with pytest.raises(ConfigNotFoundError):
        get_project_dir()


def test_load_project_spec(temp_workdir):
    with open("replicate.yaml", "w") as f:
        f.write("repository: file://.replicate/")

    os.mkdir(".replicate")
    with open(".replicate/repository.json", "w") as f:
        f.write(
            """{
  "version": 1234
}"""
        )

    project = Project()
    assert project._load_project_spec() == ProjectSpec(version=1234)


def test_load_missing_project_spec(temp_workdir):
    with open("replicate.yaml", "w") as f:
        f.write("repository: file://.replicate/")

    project = Project()
    assert project._load_project_spec() is None


def test_load_corrupted_project_spec(temp_workdir):
    with open("replicate.yaml", "w") as f:
        f.write("repository: file://.replicate/")

    project = Project()
    os.mkdir(".replicate")

    with open(".replicate/repository.json", "w") as f:
        f.write(
            """{
  "version": asdf
}"""
        )

    with pytest.raises(CorruptedProjectSpec):
        project._load_project_spec()

    with open(".replicate/repository.json", "w") as f:
        f.write(
            """{
  "foo": "bar"
}"""
        )

    with pytest.raises(CorruptedProjectSpec):
        project._load_project_spec()


def test_write_project_spec(temp_workdir):
    with open("replicate.yaml", "w") as f:
        f.write("repository: file://.replicate/")

    project = Project()
    project._write_project_spec(version=1234)

    with open(".replicate/repository.json") as f:
        assert (
            f.read()
            == """{
  "version": 1234
}"""
        )


def test_write_load_project_spec(temp_workdir):
    with open("replicate.yaml", "w") as f:
        f.write("repository: file://.replicate/")

    project = Project()
    project._write_project_spec(version=1234)
    assert project._load_project_spec().version == 1234
