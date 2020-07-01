import json
import os
import tempfile
import unittest

import replicate
from replicate.storage import DiskStorage


class TestExperiment(unittest.TestCase):
    def test_init(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            storage = DiskStorage(root=tmpdir)
            experiment = replicate.init(storage=storage, workdir=None, params={"learning_rate": 0.002})
            self.assertEqual(len(experiment.id), 64)
            metadata = json.loads(storage.get("experiments/{}/replicate-metadata.json".format(experiment.id)))
            self.assertEqual(metadata["id"], experiment.id)
            self.assertEqual(metadata["params"], {"learning_rate": 0.002})


    def test_commit(self):
        # TODO: py.test has a nice decorator for this tempdir crap
        with tempfile.TemporaryDirectory() as workdir:
            with open(os.path.join(workdir, "train.py"), "w") as fh:
                fh.write("print(1 + 1)")

            with tempfile.TemporaryDirectory() as storage_dir:
                storage = DiskStorage(root=storage_dir)
                experiment = replicate.init(storage=storage, workdir=workdir, params={"learning_rate": 0.002})
                commit = experiment.commit({"validation_loss": 0.123})
                self.assertEqual(len(commit.id), 64)
                metadata = json.loads(storage.get("commits/{}/replicate-metadata.json".format(commit.id)))
                self.assertEqual(metadata["id"], commit.id)
                self.assertEqual(metadata["experiment"]["id"], experiment.id)

                self.assertEqual(storage.get("commits/{}/train.py".format(commit.id)), "print(1 + 1)")
