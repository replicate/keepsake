import time
import json
import sys
import subprocess
import os
import pytest

from .utils import get_env


def test_gpu_training(gpu_instance, tmpdir, temp_bucket):
    tmpdir = str(tmpdir)

    # run the two tests in sequence, to save time setting up instances
    run_gpu_test("s3://" + temp_bucket, gpu_instance, tmpdir, temp_bucket)
    run_gpu_test("gs://" + temp_bucket, gpu_instance, tmpdir, temp_bucket)


def run_gpu_test(storage, gpu_instance, tmpdir, temp_bucket):
    with open(os.path.join(tmpdir, "replicate.yaml"), "w") as f:
        f.write(
            """
storage: {storage}
""".format(
                storage=storage
            )
        )
    with open(os.path.join(tmpdir, "requirements.txt"), "w") as f:
        f.write(
            """
torch==1.4.0
"""
        )
    with open(os.path.join(tmpdir, "train.py"), "w") as f:
        f.write(
            """
import replicate
import torch
import time

def main():
    experiment = replicate.init(path=".")
    num_gpus = torch.cuda.device_count()
    time.sleep(1)
    experiment.checkpoint(path=".", step=1, metrics={"num_gpus": num_gpus})

if __name__ == "__main__":
    main()
"""
        )

    env = get_env()
    env["AWS_ACCESS_KEY_ID"] = gpu_instance.aws_access_key_id
    env["AWS_SECRET_ACCESS_KEY"] = gpu_instance.aws_secret_access_key

    subprocess.run(
        [
            "replicate",
            "run",
            "-v",
            "-H",
            gpu_instance.username + "@" + gpu_instance.public_ip,
            "-i",
            gpu_instance.ssh_private_key,
            "train.py",
        ],
        cwd=tmpdir,
        env=env,
        check=True,
    )

    experiments = json.loads(
        subprocess.run(
            ["replicate", "list", "--json"],
            cwd=tmpdir,
            env=env,
            capture_output=True,
            check=True,
        ).stdout
    )
    assert len(experiments) == 1

    exp = experiments[0]
    latest = exp["latest_checkpoint"]
    assert latest["metrics"]["num_gpus"] == 1
    assert exp["running"]

    running = json.loads(
        subprocess.run(
            ["replicate", "ps", "--json"],
            cwd=tmpdir,
            env=env,
            capture_output=True,
            check=True,
        ).stdout
    )

    assert running == experiments

    time.sleep(31)  # TODO(andreas): speed this up to make CI faster
    experiments = json.loads(
        subprocess.run(
            ["replicate", "list", "--json"],
            cwd=tmpdir,
            env=env,
            capture_output=True,
            check=True,
        ).stdout
    )
    assert len(experiments) == 1

    exp = experiments[0]
    assert not exp["running"]

    running = json.loads(
        subprocess.run(
            ["replicate", "ps", "--json"],
            cwd=tmpdir,
            env=env,
            capture_output=True,
            check=True,
        ).stdout
    )
    assert len(running) == 0
