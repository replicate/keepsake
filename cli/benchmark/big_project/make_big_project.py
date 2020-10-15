import argparse
import os
import tempfile

from replicate.project import Project
from replicate.storage import storage_for_url

parser = argparse.ArgumentParser(
    description="Create two projects: one with lots of metadata, and another which is the same but with a few new projects and checkpoints to test incremental updates"
)
parser.add_argument("bucket")
parser.add_argument("bucket_prime")
args = parser.parse_args()

with tempfile.TemporaryDirectory() as project_dir:
    print("Creating project...")
    project = Project(dir=project_dir)
    for i in range(1000):
        if i % 10 == 0:
            print("Experiment", i)
        experiment = project.experiments.create(
            path=None, params={"foo": "bar"}, quiet=True
        )
        for j in range(100):
            experiment.checkpoint(path=None, metrics={"loss": 0.00001}, quiet=True)

    print("Uploading to bucket...")
    storage = storage_for_url(args.bucket)
    storage.put_path(os.path.join(project_dir, ".replicate/storage/"), "")

    print("Creating extra data...")
    for i in range(10):
        experiment.checkpoint(path=None, metrics={"loss": 0.00001}, quiet=True)
    for i in range(10):
        experiment = project.experiments.create(
            path=None, params={"foo": "bar"}, quiet=True
        )
        for j in range(100):
            experiment.checkpoint(path=None, metrics={"loss": 0.00001}, quiet=True)

    print("Uploading to bucket_prime...")
    storage = storage_for_url(args.bucket_prime)
    storage.put_path(os.path.join(project_dir, ".replicate/storage/"), "")
