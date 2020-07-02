import functools
import json
import os
import tempfile
import unittest

import replicate
from replicate.storage import DiskStorage


def temp_workdir(f):
    @functools.wraps(f)
    def wrapper(*args, **kwargs):
        orig_cwd = os.getcwd()
        try:
            with tempfile.TemporaryDirectory() as tmpdir:
                os.chdir(tmpdir)
                return f(*args, **kwargs)
        finally:
            os.chdir(orig_cwd)
    return wrapper


class TestExperiment(unittest.TestCase):
    @temp_workdir
    def test_init(self):
        experiment = replicate.init(params={"learning_rate": 0.002})

        self.assertEqual(len(experiment.id), 64)
        with open(".replicate/storage/experiments/{}/replicate-metadata.json".format(experiment.id)) as fh:
            metadata = json.load(fh)
        self.assertEqual(metadata["id"], experiment.id)
        self.assertEqual(metadata["params"], {"learning_rate": 0.002})


    @temp_workdir
    def test_commit(self):
        with open("train.py", "w") as fh:
            fh.write("print(1 + 1)")

        experiment = replicate.init(params={"learning_rate": 0.002})
        commit = experiment.commit({"validation_loss": 0.123})

        self.assertEqual(len(commit.id), 64)
        with open(".replicate/storage/commits/{}/replicate-metadata.json".format(commit.id)) as fh:
            metadata = json.load(fh)
        self.assertEqual(metadata["id"], commit.id)
        self.assertEqual(metadata["metrics"], {"validation_loss": 0.123})
        self.assertEqual(metadata["experiment"]["id"], experiment.id)

        with open(".replicate/storage/commits/{}/train.py".format(commit.id)) as fh:
            self.assertEqual(fh.read(), "print(1 + 1)")
