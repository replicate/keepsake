import socket
import time
import string
import random
import pytest
import boto3
from botocore.exceptions import ClientError, NoCredentialsError


@pytest.fixture(scope="function")
def temp_bucket():
    # We don't create bucket here so we can test Replicate's ability to create
    # buckets.

    bucket_name = "replicate-test-" + "".join(
        random.choice(string.ascii_lowercase) for _ in range(20)
    )
    yield bucket_name

    # FIXME: this is used for both GCP and S3 tests, but only cleans up
    # the S3 bucket.
    # These fixtures should probably be arranged more intelligently so there
    # are different ones for GCP and S3.
    try:
        s3 = boto3.resource("s3")
        bucket = s3.Bucket(bucket_name)
        bucket.objects.all().delete()
        bucket.delete()
    except (NoCredentialsError, ClientError):
        pass


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
