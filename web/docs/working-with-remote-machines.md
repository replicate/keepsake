---
id: working-with-remote-machines
title: Working with remote machines
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

This guide is the second part of the tutorial. If you want to follow along with the commands here, you might want to do [the first part](tutorial.md).

Replicate makes it really easy to work with multiple training machines. It lets you:

1. **Store your experiment data in the cloud on Google Cloud Storage or Amazon S3.** This means you can store results from multiple training machines in one place and collaborate with other people.
2. **Run a training job on a remote machine.** Replicate copies your code to a remote server, sets up a reproducible environment inside Docker, then runs your training script inside it.

## Store data on cloud storage

By default, Replicate stores your experiments and commits in `.replicate/storage/` in the same directory as `replicate.yaml` on your local disk.

This means you can run Replicate on a training instance in the cloud or on real hardware, and the results will be stored in a central location. When you run `replicate ls` on your local machine, it will list all of the experiments that have run on any training machine, so you can easily compare between them and checkout the results.

<Tabs
groupId="cloud-service"
defaultValue="gcp"
values={[
{label: 'Google Cloud Storage', value: 'gcp'},
{label: 'Amazon S3', value: 'aws'},
]
}>
<TabItem value="gcp">

To store data on Google Cloud Storage, you first need to [install the Google Cloud SDK](https://cloud.google.com/sdk/docs) if you haven't already. Then, run `gcloud init` and follow its instructions:

- Log in to your Google account when prompted.
- If it asks to choose a project, pick the default or first option.
- If it asks to pick a default region, hit enter and select `1` (you want a default region, but it doesn't matter which one).

Next, run this, because Cloud SDK needs you to log in a second time to let other applications use your Google Cloud account:

```shell-session
gcloud auth application-default login
```

Then, create a file called `replicate.yaml` in the same directory as your project with this content, replacing `[username]` with your name, or some other unique string:

```yaml
storage: "gs://replicate-[username]-iris-classifier"
```

</TabItem>
<TabItem value="aws">

We haven't written a full guide to setting up S3 yet, so this assumes some knowledge of how Amazon Web Services works.

To store data on Amazon S3, first set the environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` with your access key and secret. You'll probably want to use an IAM user with limited permissions.

Then, create `replicate.yaml` with this content, replacing `[username]` with your name, or some other unique string:

```yaml
storage: "s3://replicate-[username]-iris-classifier"
```

</TabItem>
</Tabs>

Now, when you run your training script, calling `experiment.commit()` will upload all your working directory and commit metadata to this Google Cloud Storage bucket.

If you're following along from [the tutorial](tutorial.md), run `python train.py` again. When you run `replicate ls`, it will list experiments from the bucket.

## Train on remote machines

A common pattern for training is to copy the code you are working on locally to a remote GPU machine, and then start the training job. Then, when it's finished, copy the results and artifacts back to your local machine or cloud storage.

Replicate makes this pattern really easy. It can run training jobs anywhere SSH where Docker is installed that you can log into via SSH.

### Create a training machine

The first step is to create a machine to run training jobs. If you're already got a machine that you can access via SSH, you're already sorted. Otherwise, you need to create a cloud GPU instance.

<Tabs
groupId="cloud-service"
defaultValue="gcp"
values={[
{label: 'Google Cloud Compute', value: 'gcp'},
{label: 'Amazon EC2', value: 'aws'},
]
}>
<TabItem value="gcp">

Run this command to create a Google Cloud Compute instance:

```shell-session
gcloud compute instances create \
    replicate-training-1 \
    --zone=us-east1-c \
    --machine-type=n1-standard-4 \
    --accelerator type=nvidia-tesla-k80,count=1 \
    --boot-disk-size=500GB \
    --image-project=deeplearning-platform-release \
    --image-family=common-cu101 \
    --maintenance-policy TERMINATE \
    --restart-on-failure \
    --scopes=default,storage-rw \
    --metadata="install-nvidia-driver=True"
```

Make a note of the value underneath `EXTERNAL_IP`. You'll need that in the next section.

:::note
If you're using Google Cloud Storage to store your data, it will only work on instances that you have started with the `--scopes=default,storage-rw` option, as we have done here

This is because we don't pass Google Cloud auth credentials to remote machines yet. We're working on fixing this – stay tuned!
:::

</TabItem>
<TabItem value="aws">

We haven't finished the Amazon EC2 section yet, sorry! Switch back to the "Google Cloud Compute" tab.

</TabItem>
</Tabs>

### Run a training job

Next, we need to tell Replicate what version of Python you want to use. Add this line to `replicate.yaml`:

```yaml
python: "3.8"
```

You can now use `replicate run` to run a training job on a remote machine. The only two requirements are the machine has Docker installed on it, and you can log into it with SSH.

Run this, replacing `EXTERNAL_IP` with the IP address of the machine:

```shell-session
$ replicate run -H EXTERNAL_IP python train.py
═══╡ Found CUDA driver on remote host: 418.87.01
═══╡ Building Docker image...
[+] Building 181.7s (12/12) FINISHED
...
═══╡ Running 'python train.py'...
Epoch 0, train loss: 1.184, validation accuracy: 0.333
Epoch 1, train loss: 1.117, validation accuracy: 0.333
Epoch 2, train loss: 1.061, validation accuracy: 0.467
...
Epoch 97, train loss: 0.121, validation accuracy: 1.000
Epoch 98, train loss: 0.119, validation accuracy: 1.000
Epoch 99, train loss: 0.118, validation accuracy: 1.000
```

When it's finished, you should see the experiment show up in `replicate ls`:

```shell-session
$ replicate ls
EXPERIMENT  STARTED            STATUS   HOST             USER  LATEST COMMIT      LOSS    BEST COMMIT        LOSS
274a9ec     3 minutes ago      stopped  34.75.189.211    ben   911cf2b (step 99)  0.1176  911cf2b (step 99)  0.1176
```

You can create as many experiments as you like on as many machines as you like, and they'll all show up in one place.

If you want to get the results of the experiment, run `replicate checkout`, like you did in [the first part of the tutorial](tutorial.md#check-out-a-commit).

## Stop your training machine

When you've finished working, remember to stop your training machine, otherwise you're going to run up a large bill!

<Tabs
groupId="cloud-service"
defaultValue="gcp"
values={[
{label: 'Google Cloud Compute', value: 'gcp'},
{label: 'Amazon EC2', value: 'aws'},
]
}>
<TabItem value="gcp">

```shell-session
gcloud compute instances delete replicate-training-1 --zone=us-east1-c
```

</TabItem>
<TabItem value="aws">
We haven't finished the Amazon EC2 section yet, sorry! Switch back to the "Google Cloud Compute" tab.
</TabItem>
</Tabs>

## What's next

You might want to take a look at:

- [Some of our example models](example-models)
- [How Replicate works under the hood](how-it-works.md)
- [The `replicate.yaml` reference](replicate-yaml.md)

If something doesn't make sense, doesn't work, or you just have some questions, please email us: [team@replicate.ai](mailto:team@replicate.ai). We love hearing from you!
