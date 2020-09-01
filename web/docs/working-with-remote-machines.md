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

By default, Replicate stores your experiments and checkpoints in `.replicate/storage/` in your working directory. You can also store this data on cloud storage, like Google Cloud Storage or Amazon S3.

This means you can train on several machines and the results will be stored in a central location. When you run `replicate ls` on your local machine, it will list all of the experiments that have run anywhere, so you can easily compare between them and download the results.

<Tabs
groupId="cloud-service"
defaultValue="gcp"
values={[
{label: 'Google Cloud Storage', value: 'gcp'},
{label: 'Amazon S3', value: 'aws'},
]
}>
<TabItem value="gcp">

### Log in to Google Cloud

To store data on Google Cloud Storage, you first need to [install the Google Cloud SDK](https://cloud.google.com/sdk/docs) if you haven't already. Then, run `gcloud init` and follow its instructions:

- Log in to your Google account when prompted.
- If it asks to choose a project, pick the default or first option.
- If it asks to pick a default region, hit enter and select `1` (you want a default region, but it doesn't matter which one).

Next, run this, because Cloud SDK needs you to log in a second time to let other applications use your Google Cloud account:

```shell-session
gcloud auth application-default login
```

### Point Replicate at Google Cloud

Create a file called `replicate.yaml` in the same directory as your project with this content, replacing `[your username]` with your name, or some other unique string:

```yaml
storage: "gs://replicate-[your username]-iris-classifier"
```

</TabItem>
<TabItem value="aws">

We haven't written a full guide to setting up S3 yet, so this assumes some knowledge of how Amazon Web Services works.

To store data on Amazon S3, first set the environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` with your access key and secret. You'll probably want to use an IAM user with limited permissions.

Then, create `replicate.yaml` with this content, replacing `[your username]` with your name, or some other unique string:

```yaml
storage: "s3://replicate-[your username]-iris-classifier"
```

</TabItem>
</Tabs>

Now, when you run your training script, calling `experiment.checkpoint()` will upload all your working directory and checkpoint metadata to this Google Cloud Storage bucket.

:::note
Replicate will automatically create the bucket for you if it doesn't exist.
:::

If you're following along from [the tutorial](tutorial.md), run `python train.py` again. This time it will save your experiments to the Google Cloud bucket. (It takes a second to save each checkpoint, so press `Ctrl-C` after a few epochs if you don't want to wait.)

Now, when you run `replicate ls`, it will list experiments from the bucket.

:::note
You will not see the experiments you ran locally in the first part of the tutorial. Replicate only lists experiments from one storage location. Now you've switched from your local storage to Google Cloud, it will no longer list your local experiments.

It is easy to migrate locations, if you ever need to, because Replicate just stores its data as plain files. In this instance, you would copy the contents of `.replicate/storage/` to `gs://replicate-[username]-iris-classifier`.
:::

## Train on remote machines

Anywhere your training script runs, Replicate will keep track of it. Now, with the `storage` option set in `replicate.yaml`, you can run your training script on a remote training machine and Replicate will keep track of it in cloud storage.

However, to do this, you may need to copy code from your local development machine to the training machine. And, you need to install Python dependencies to make it work. And, you might need to wrestle with CUDA drivers.

Replicate has a shortcut that makes this really easy. It will copy the code onto the training machine, build a Docker image to create a reproducible environment for your training, then start the training script inside. It works on any machine with SSH and Docker.

### Create a training machine

The first step is to create a machine to run training jobs. If you're already got a machine that you can access via SSH, you're good to go. Otherwise, you need to create a cloud GPU instance.

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

You can now use `replicate run` to run a training job on a remote machine. The only two requirements are that the machine has Docker installed on it, and you can log into it with SSH.

Run this on your local machine, replacing `EXTERNAL_IP` with the IP address of the machine:

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

:::note
If there is a "connection refused" error, wait a few minutes. The Google Cloud server takes some time to start up.
:::

Replicate copies the code from your local machine to the training server, builds a Docker image with the correct CUDA version, then starts your training script.

When it's finished, run `replicate ls` on your local machine, and you should see the experiment from the training server show up:

```shell-session
$ replicate ls
EXPERIMENT  STARTED            STATUS   HOST             USER  LATEST CHECKPOINT  LOSS    BEST CHECKPOINT    LOSS
274a9ec     3 minutes ago      stopped  34.75.189.211    ben   911cf2b (step 99)  0.1176  911cf2b (step 99)  0.1176
```

Because you're storing the experiment data on cloud storage, you can create as many experiments as you like on as many machines as you like, and they'll all show up in one place.

If you want to get the results of the experiment, run `replicate checkout`, like you did in [the first part of the tutorial](tutorial.md#check-out-a-checkpoint).

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

<!-- - [Some of our example models](example-models) -->

- [How Replicate works under the hood](how-it-works.md)
- [The `replicate.yaml` reference](replicate-yaml.md)

If something doesn't make sense, doesn't work, or you just have some questions, please email us: [team@replicate.ai](mailto:team@replicate.ai). We love hearing from you!
