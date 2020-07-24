import time
import json
import sys
import subprocess
import os
import pytest


def test_gpu_training(gpu_instance, tmpdir, temp_bucket):
    tmpdir = str(tmpdir)

    storage = "s3://" + temp_bucket

    with open(os.path.join(tmpdir, "replicate.yaml"), "w") as f:
        f.write(
            """
storage: {storage}
""".format(
                storage=storage
            )
        )
    with open(os.path.join(tmpdir, "requirements.txt"), "w") as f:
        f.write("torch==1.4.0")
    with open(os.path.join(tmpdir, "train.py"), "w") as f:
        f.write(
            """
import replicate
import torch
import time

def main():
    experiment = replicate.init()
    num_gpus = torch.cuda.device_count()
    time.sleep(1)
    experiment.commit(metrics={"num_gpus": num_gpus})

if __name__ == "__main__":
    main()
"""
        )

    env = os.environ
    env["PATH"] = "/usr/local/bin:" + os.environ["PATH"]
    env["AWS_ACCESS_KEY_ID"] = gpu_instance.aws_access_key_id
    env["AWS_SECRET_ACCESS_KEY"] = gpu_instance.aws_secret_access_key

    proc = subprocess.run(
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
        stdout=sys.stdout,
        stderr=sys.stderr,
    )
    assert proc.returncode == 0

    experiments = json.loads(
        subprocess.check_output(
            ["replicate", "list", "--format=json"], cwd=tmpdir, env=env,
        )
    )
    assert len(experiments) == 1

    exp = experiments[0]
    latest = exp["latest_commit"]
    assert latest["metrics"] == {"num_gpus": 1}
    assert exp["running"]

    time.sleep(31)  # TODO(andreas): speed this up to make CI faster
    experiments = json.loads(
        subprocess.check_output(
            ["replicate", "list", "--format=json"], cwd=tmpdir, env=env,
        )
    )
    assert len(experiments) == 1

    exp = experiments[0]
    assert not exp["running"]
