---
id: tutorial
title: First steps
---

import CodeBlock from "@theme/CodeBlock";
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import config from '../docusaurus.config.js';

If you like to **learn by doing**, this guide will help you learn how Replicate works by building a simple model.

If you prefer to **learn concepts first**, take a look at [our guide about how Replicate works](how-it-works).

## Install Replicate

<Tabs
groupId="operating-systems"
defaultValue="mac"
values={[
{label: 'macOS', value: 'mac'},
{label: 'Linux', value: 'linux'},
]
}>
<TabItem value="mac">

Run the following commands in a terminal:

```shell-session
curl -o /usr/local/bin/replicate https://storage.googleapis.com/replicate-public/cli/latest/darwin/amd64/replicate
chmod +x /usr/local/bin/replicate
```

</TabItem>
<TabItem value="linux">
Run the following commands in a terminal:

```shell-session
sudo curl -o /usr/local/bin/replicate https://storage.googleapis.com/replicate-public/cli/latest/linux/amd64/replicate
sudo chmod +x /usr/local/bin/replicate
```

</TabItem>
</Tabs>

## Write a model

We're going to make a model that classifies Iris plants, trained on the [Iris dataset](https://archive.ics.uci.edu/ml/datasets/iris). It's an intentionally simple model that trains really fast, just so we can show you how Replicate works.

First, let's make a directory to work in:

```shell-session
mkdir iris-classifier
cd iris-classifier
```

Then, copy and paste this code into `train.py`:

<CodeBlock className="python" metastring='title="train.py"'>
{require('!!raw-loader!../../example/train.py').default}
</CodeBlock>
 

Notice there are two highlighted lines that call Replicate. They don't affect the behavior of the training – they just save data in Replicate to keep track of what is going on.

The first is `replicate.init()`. This creates an **experiment**, which represents a run of your training script. The experiment records the hyperparameters you pass to it and makes a copy of your code.

The second is `experiment.checkpoint()`. This creates a **checkpoint** within the experiment. The checkpoint saves the metrics at that point, and makes a copy of the file or directory you pass to it, which could include weights and any other artifacts.

**Each experiment contains multiple checkpoints.** You typically save your model periodically during training, because the best result isn't necessarily the most recent one. A checkpoint is created just after you save your model, so Replicate can keep track of versions of your saved model.

## Install the dependencies

Before we start training, we need to install the Python packages that the model needs.

Create `requirements.txt` to define our requirements:

<CodeBlock className="txt" metastring='title="requirements.txt"'>
{`replicate==`+config.customFields.version+`
scikit-learn==0.23.1
torch==1.4.0
`}</CodeBlock>
 

Then, install the Python requirements inside a [Virtualenv](https://virtualenv.pypa.io/en/latest/):

```shell-session
virtualenv venv
source venv/bin/activate
pip install -r requirements.txt
```

## Train the model

We're now going to train this model a couple of times with different parameters to see what we can do with Replicate.

First, train it with default parameters:

```shell-session
$ python train.py
Epoch 0, train loss: 1.184, validation accuracy: 0.333
Epoch 1, train loss: 1.117, validation accuracy: 0.333
Epoch 2, train loss: 1.061, validation accuracy: 0.467
...
Epoch 97, train loss: 0.121, validation accuracy: 1.000
Epoch 98, train loss: 0.119, validation accuracy: 1.000
Epoch 99, train loss: 0.118, validation accuracy: 1.000
```

Next, run the training with a different learning rate:

```shell-session
$ python train.py --learning_rate=0.2
Epoch 0, train loss: 1.184, validation accuracy: 0.333
Epoch 1, train loss: 1.161, validation accuracy: 0.633
Epoch 2, train loss: 1.124, validation accuracy: 0.667
...
Epoch 97, train loss: 0.057, validation accuracy: 0.967
Epoch 98, train loss: 0.057, validation accuracy: 0.967
Epoch 99, train loss: 0.056, validation accuracy: 0.967
```

## Experiments and checkpoints

The calls to the `replicate` Python library have saved your experiments locally. You can use `replicate ls` to list them:

```shell-session
$ replicate ls
EXPERIMENT  STARTED         STATUS   HOST             USER  LEARNING_RATE  LATEST CHECKPOINT  ACCURACY  BEST CHECKPOINT    ACCURACY
f1b4d2b     13 seconds ago  stopped  107.133.144.125  ben   0.01           8ee945b (step 99)  1         8ee945b (step 99)  1
02010f9     4 seconds ago   stopped  107.133.144.125  ben   0.2            dbcdd99 (step 99)  0.96667   2a2ed68 (step 12)  1
```

:::note
Similar to how Git works, all this data is in your working directory. Replicate only lists experiments in the current working directory, so you'll only see experiments from the model you're working on.

If you want to poke around at the internal data, it is inside `.replicate/`. It's not something you'd do day to day, but there's no magic going on – everything's right there as files and plain JSON.

You probably want to add `.replicate` to `.gitignore`, if you're using Git.
:::

As a reminder, this is a list of **experiments** which represents runs of the `train.py` script. They store a copy of the code as it was when the script was started.

Within experiments are **checkpoints**, which are created every time you call `experiment.checkpoint()` in your training script. The checkpoint contains your weights, Tensorflow logs, and any other artifacts you want to save.

To list the checkpoints within an experiment, you can use `replicate show`. Run this, replacing `f1b4d2b` with an experiment ID from your output of `replicate ls`:

```shell-session
$ replicate show f1b4d2b
Experiment: f1b4d2b38044f52721c907300add177c3e12c7c6e90307b2312da8d75de1f494

Created:        Tue, 01 Sep 2020 21:42:22 PDT
Status:         stopped
Host:           107.133.144.125
User:           ben
Command:        train.py

Params
learning_rate:  0.01
num_epochs:     100

Checkpoints
ID       STEP  CREATED             ACCURACY  LOSS
109b73b  0     about a minute ago  0.33333   1.1836
156d2dd  1     about a minute ago  0.33333   1.1173
612f909  2     about a minute ago  0.46667   1.0611
ac322e7  3     about a minute ago  0.63333   1.0138
ebfdb6f  4     about a minute ago  0.7       0.97689
...
d3916c4  95    about a minute ago  1         0.12417
a109647  96    about a minute ago  1         0.12244
ba5f308  97    about a minute ago  1         0.12076
4c49656  98    about a minute ago  1         0.11915
8ee945b  99    about a minute ago  1 (best)  0.1176
```

You can also use `replicate show` on a checkpoint to get all the information about it. Run this, replacing `8ee` with a checkpoint ID from the experiment:

```shell-session
$ replicate show 8ee
Checkpoint: 8ee945b289bb762e81027db825eb0d7d9e4b0d0a3001f6cb8ceb6aec5c00c089

Created:            Tue, 01 Sep 2020 21:42:23 PDT
Path:               model.pth
Step:               99

Experiment
ID:                 f1b4d2b38044f52721c907300add177c3e12c7c6e90307b2312da8d75de1f494
Created:            Tue, 01 Sep 2020 21:42:22 PDT
Status:             stopped
Host:               107.133.144.125
User:               ben
Command:            train.py

Params
learning_rate:      0.01
num_epochs:         100

Metrics
accuracy:           1 (primary, maximize)
loss:               0.11759971082210541
```

:::note
Notice you can pass a prefix to `replicate show`, and it'll automatically find the experiment that starts with just those characters. Saves a few keystrokes.
:::

## Compare checkpoints

Let's compare the last checkpoints from the two experiments we ran. Run this, replacing `8ee945b` and `dbcdd99` with the two checkpoint IDs from the `LATEST CHECKPOINT` column in `replicate ls`:

```shell-session
$ replicate diff 8ee945b dbcdd99
Checkpoint:               8ee945b                   dbcdd99
Experiment:               f1b4d2b                   02010f9

Params
learning_rate:            0.01                      0.2

Metrics
accuracy:                 1                         0.9666666388511658
loss:                     0.11759971082210541       0.056485891342163086
```

`replicate diff` works a bit like `git diff`, except in addition to the code, it compares all of the metadata that Replicate is aware of: params, metrics, dependencies, and so on.

:::note
`replicate diff` compares **checkpoints**, because that is the thing that actually has all the results.

You can also pass an experiment ID, and it will pick the best or latest checkpoint from that experiment.
:::

## Check out a checkpoint

At some point you might want to get back to some point in the past. Maybe you've run a bunch of experiments in parallel, and you want to choose one that works best. Or, perhaps you've gone down a line of exploration and it's not working, so you want to get back to where you were a week ago.

The `replicate checkout` command will copy the code and weights from a checkpoint into your working directory. Run this, replacing `d4fb0d3` with a checkpoint ID you passed to `replicate diff`:

```shell-session
$ replicate checkout 8ee945b
═══╡ The directory "/Users/ben/p/tmp/iris-classifier" is not empty.
═══╡ This checkout may overwrite existing files. Make sure you've committed everything to Git so it's safe!

Do you want to continue? (y/N) y

═══╡ Checked out d4fb0d3 to "/Users/ben/p/tmp/iris-classifier"
```

The model file in your working directory is now the model saved in that checkpoint:

```shell-session
$ ls -lh model.pth
-rw-r--r--  1 ben  staff   8.3K Aug  7 16:42 model.pth
```

This is useful for getting the trained model out of a checkpoint, but **it also copies all of the code from the experiment that checkpoint is part of**. If you made a change to the code and didn't commit to Git, `replicate checkout` will allow you get back the exact code from an experiment.

**This means you don't have to remember to commit to Git when you're running experiments.** Just try a bunch of things, then when you've found something that works, use Replicate to get back to the exact code that produced those results and formally commit it to Git.

Neat, huh? Replicate is keeping track of everything in the background so you don't have to.

## The workflow so far

With these tools, let's recap what the workflow looks like:

- Add `experiment = replicate.init()` and `experiment.checkpoint()` to your training code.
- Run several experiments by running the training script as usual, with changes to the hyperparameters or code.
- See the results of our experiments with `replicate ls` and `replicate show`.
- Compare the differences between experiments with `replicate diff`.
- Get the code from the best experiment with `replicate checkout`.
- Commit that code cleanly to Git.

You don't have to keep track of what you changed in your experiments, because Replicate does that automatically for you. You can also safely change things without committing to Git, because `replicate checkout` will always be able to get you back to the exact environment the experiment was run in.

## What's next

So far, everything we've been doing has been local on a single machine. But in practice, you probably want to train on a separate machine with GPUs, or perhaps multiple machines at the same time. Or, perhaps you want to share your experiments with other people in your team.

Take a look at [the guide to working with remote machines](working-with-remote-machines.md) to learn more about this.
