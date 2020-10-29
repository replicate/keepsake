import os
from pathlib import Path
import boto3
import botocore
from google.cloud import storage as google_storage

ROOT_DIRECTORY = Path(__file__).parent.parent.parent


def get_env():
    """
    Returns environment for running Replicate commands in
    """
    env = os.environ
    env["PATH"] = "/usr/local/bin:" + os.environ["PATH"]
    return env


def path_exists(repository, path):
    """
    Check that a path exists in a specific repository.
    repository has the format backend://root(/parent_folder),
    where backend can be file, s3, or gs.
    """
    backend, root = repository.split("://")
    assert backend in ("file", "s3", "gs")

    if backend == "file":
        return (Path(root) / path).exists()

    # append any parent_folder that's part of the repository string
    # to the path, so that root is just the bucket
    root = Path(root)
    if len(root.parts) > 1:
        path = Path(*root.parts[1:]) / path
    root = root.parts[0]

    if backend == "s3":
        s3 = boto3.resource("s3")
        try:
            s3.Object(root, str(path)).load()
            return True
        except botocore.exceptions.ClientError as e:
            if e.response['Error']['Code'] == "404":
                return False
            else:
                # Something else has gone wrong.
                raise
    if backend == "gs":
        storage_client = google_storage.Client()
        bucket = storage_client.bucket(root)
        blob = bucket.blob(str(path))
        return blob.exists()
