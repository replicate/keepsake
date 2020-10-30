import tempfile
import os
import pytest


@pytest.fixture
def temp_workdir():
    orig_cwd = os.getcwd()
    try:
        with tempfile.TemporaryDirectory() as tmpdir:
            os.chdir(tmpdir)
            yield tmpdir
    finally:
        os.chdir(orig_cwd)
