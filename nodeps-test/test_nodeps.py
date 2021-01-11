import tempfile
import shutil
import os
import unittest  # using unittest since it's in python's stdlib

import replicate


class TestPythonNoDependencies(unittest.TestCase):
    def setUp(self):
        self.test_dir = tempfile.mkdtemp()
        self.cwd = os.getcwd()
        os.chdir(self.test_dir)

    def tearDown(self):
        os.chdir(self.cwd)
        shutil.rmtree(self.test_dir)

    def test_end_to_end(self):
        with open("replicate.yaml", "w") as f:
            f.write('repository: "file://.replicate"')

        with open("foo.txt", "w") as f:
            f.write("foo")
        with open("bar.txt", "w") as f:
            f.write("bar")

        experiment = replicate.init(path=".", params={"myint": 10, "myfloat": 0.1})

        with open("bar.txt", "w") as f:
            f.write("barrrr")

        experiment.checkpoint(path="bar.txt", metrics={"value": 123.45})

        experiment = replicate.experiments.get(experiment.id)
        self.assertEqual(10, experiment.params["myint"])
        self.assertEqual(0.1, experiment.params["myfloat"])
        self.assertEqual(123.45, experiment.checkpoints[0].metrics["value"])

        foo = experiment.checkpoints[0].open("foo.txt")
        self.assertEqual("foo", foo.read().decode("utf-8"))
        bar = experiment.checkpoints[0].open("bar.txt")
        self.assertEqual("barrrr", bar.read().decode("utf-8"))

        with self.assertRaises(ImportError):
            experiment.plot("value")
