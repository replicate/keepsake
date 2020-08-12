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

Replicate can run training jobs anywhere accessible via SSH where Docker is installed

## What's next

You might want to take a look at the [`replicate.yaml` reference](replicate-yaml.md), or take a look at some of our [example models](example-models).

If something doesn't make sense, doesn't work, or you just have some questions, please email us: [team@replicate.ai](mailto:team@replicate.ai). We love hearing from you!

```

```
