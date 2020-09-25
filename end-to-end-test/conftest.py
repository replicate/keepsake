import socket
import time
from dataclasses import dataclass
import string
import random
import subprocess
import pytest
import paramiko
import boto3
from botocore.config import Config
from botocore.exceptions import NoCredentialsError
from mypy_boto3_ec2 import EC2ServiceResource
from mypy_boto3_ec2.service_resource import Instance


@dataclass
class TempInstance:
    public_ip: str
    ssh_private_key: str
    username: str
    aws_access_key_id: str
    aws_secret_access_key: str

    def exec_output(self, cmdline):
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        ssh.connect(
            self.public_ip_address,
            username=self.username,
            key_filename=self.ssh_private_key,
        )

        stdin, stdout, stderr = ssh.exec_command(cmdline)
        return stdout.read(), stderr.read()


def pytest_addoption(parser):
    parser.addoption("--aws-access-key-id", action="store")
    parser.addoption("--aws-secret-access-key", action="store")
    parser.addoption("--ssh-private-key", action="store")


@pytest.fixture(scope="function")
def gpu_instance(request):
    aws_access_key_id = request.config.getoption("--aws-access-key-id")
    aws_secret_access_key = request.config.getoption("--aws-secret-access-key")
    ssh_private_key = request.config.getoption("--ssh-private-key")
    ami_id = "ami-0800ac2bbf9b818db"  # deep learning AMI (ami-0507b1ad768ac009d) + us.gcr.io/replicate/base-ubuntu18.04-python3.7-cuda10.1-cudnn7-pytorch1.4.0:0.3 already pulled
    username = "ubuntu"

    session = boto3.Session(
        region_name="us-east-1",
        aws_access_key_id=aws_access_key_id,
        aws_secret_access_key=aws_secret_access_key,
    )

    print("Creating instance")

    ec2: EC2ServiceResource = session.resource("ec2")
    # TODO(andreas): run ci instance in separate subnet for isolation?
    instance: Instance = ec2.create_instances(
        ImageId=ami_id, MaxCount=1, MinCount=1, InstanceType="p2.xlarge", KeyName="ci",
    )[0]
    try:
        instance.wait_until_running()
        instance.reload()

        print("Waiting for instance to become accessible")
        wait_for_port(host=instance.public_ip_address, port=22, timeout=120.0)

        yield TempInstance(
            public_ip=instance.public_ip_address,
            ssh_private_key=ssh_private_key,
            username=username,
            aws_access_key_id=aws_access_key_id,
            aws_secret_access_key=aws_secret_access_key,
        )
    finally:
        instance.terminate()
        # save some time by not calling
        # instance.wait_until_terminated()


@pytest.fixture(scope="function")
def temp_bucket():
    # FIXME(bfirsh): this seems to not pass access key like gpu_instance?
    s3 = boto3.resource("s3")
    bucket_name = "replicate-test-" + "".join(
        random.choice(string.ascii_lowercase) for _ in range(20)
    )

    try:
        s3.create_bucket(Bucket=bucket_name)
        bucket = s3.Bucket(bucket_name)
        bucket.wait_until_exists()
        yield bucket_name
        bucket = s3.Bucket(bucket_name)
        bucket.objects.all().delete()
        bucket.delete()
    # when just doing local stuff, not being able to create a bucket is fine.
    except NoCredentialsError:
        yield bucket_name


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
