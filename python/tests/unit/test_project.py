import os
import tempfile

import pytest

from replicate.project import get_project_dir


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

    # default case: working directory
    assert get_project_dir() == os.path.join(root, "foo/bar")

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
