import socket
import time
import string
import random
import pytest
import boto3
from botocore.exceptions import ClientError, NoCredentialsError
from google.cloud import storage as google_storage


class TempBucketFactory:
    def __init__(self):
        self.s3_bucket_names = []
        self.gs_bucket_names = []

    def make_name(self):
        return "keepsake-test-endtoend-" + "".join(
            random.choice(string.ascii_lowercase) for _ in range(20)
        )

    def s3(self):
        name = self.make_name()
        self.s3_bucket_names.append(name)
        return name

    def gs(self):
        name = self.make_name()
        self.gs_bucket_names.append(name)
        return name

    def cleanup(self):
        if self.s3_bucket_names:
            s3 = boto3.resource("s3")
            for bucket_name in self.s3_bucket_names:
                bucket = s3.Bucket(bucket_name)
                bucket.objects.all().delete()
                bucket.delete()
        if self.gs_bucket_names:
            storage_client = google_storage.Client()
            for bucket_name in self.gs_bucket_names:
                bucket = storage_client.bucket(bucket_name)
                for blob in bucket.list_blobs():
                    blob.delete()
                bucket.delete()


@pytest.fixture(scope="function")
def temp_bucket_factory() -> TempBucketFactory:
    # We don't create bucket here so we can test Keepsake's ability to create
    # buckets.
    factory = TempBucketFactory()
    yield factory
    factory.cleanup()


def wait_for_port(port, host="localhost", timeout=5.0):
    """Wait until a port starts accepting TCP connections.
    Args:
        port (int): Port number.
        host (str): Host address on which the port should exist.
        timeout (float): In seconds. How long to wait before raising errors.
    Raises:
        TimeoutError: The port isn't accepting connection after time specified in `timeout`.
    """
    start_time = time.perf_counter()
    while True:
        try:
            with socket.create_connection((host, port), timeout=timeout):
                break
        except OSError as ex:
            time.sleep(1.0)
            if time.perf_counter() - start_time >= timeout:
                raise TimeoutError(
                    "Waited too long for the port {} on host {} to start accepting connections.".format(
                        port, host
                    )
                ) from ex
