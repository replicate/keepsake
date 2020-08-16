---
id: tutorial
title: First steps
---

import CodeBlock from "@theme/CodeBlock";
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import config from '../docusaurus.config.js';

This guide will help you learn how Replicate works by building a simple model.

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

```python title="train.py" {12,52}
import argparse
import replicate
from sklearn.datasets import load_iris
from sklearn.model_selection import train_test_split
from sklearn.utils import shuffle
import torch
from torch import nn
from torch.autograd import Variable


def train(learning_rate, num_epochs):
    experiment = replicate.init(learning_rate=learning_rate, num_epochs=num_epochs)

    iris = load_iris()
    train_features, val_features, train_labels, val_labels = train_test_split(
        iris.data,
        iris.target,
        train_size=0.8,
        test_size=0.2,
        random_state=0,
        stratify=iris.target,
    )
    train_features = torch.FloatTensor(train_features)
    val_features = torch.FloatTensor(val_features)
    train_labels = torch.LongTensor(train_labels)
    val_labels = torch.LongTensor(val_labels)

    torch.manual_seed(0)
    model = nn.Sequential(nn.Linear(4, 15), nn.ReLU(), nn.Linear(15, 3),)
    optimizer = torch.optim.Adam(model.parameters(), lr=learning_rate)
    criterion = nn.CrossEntropyLoss()

    for epoch in range(num_epochs):
        model.train()
        optimizer.zero_grad()
        outputs = model(train_features)
        loss = criterion(outputs, train_labels)
        loss.backward()
        optimizer.step()

        with torch.no_grad():
            model.eval()
            output = model(val_features)
            accuracy = (output.argmax(1) == val_labels).float().sum() / len(val_labels)

        print(
            "Epoch {}, train loss: {:.3f}, validation accuracy: {:.3f}".format(
                epoch, loss.item(), accuracy
            )
        )
        torch.save(model, "model.pth")
        experiment.commit(step=epoch, loss=loss.item(), accuracy=accuracy)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--learning_rate", type=float, default=0.01)
    parser.add_argument("--num_epochs", type=int, default=100)
    args = parser.parse_args()
    train(args.learning_rate, args.num_epochs)
```

Notice there are two highlighted lines that call Replicate. They don't affect the behavior of the training – they just save data in Replicate to keep track of what is going on.

The first is `replicate.init()`. This creates an **experiment**, which represents the training run. This is called once at the start so you can pass your hyperparameters.

The second is `experiment.commit()`. This creates a **commit** within the experiment, which saves the filesystem at that point and any metrics you pass to the function.

**Each commit is a complete snapshot of the filesystem at that point.** All your code, weights, Tensorboard logs, and so on are all saved. This means once you've found that precise checkpoint that produces the best results, you can get back to that point.

**Each experiment contains multiple commits.** You typically save your model periodically during training, because the best result isn't necessarily the most recent one. You commit to Replicate just after you save your model, so it can keep track of these versions for you.

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

## Experiments and commits

The calls to the `replicate` Python library have saved your experiments locally. You can use `replicate list` to list them:

```shell-session
$ replicate list
EXPERIMENT  STARTED         STATUS   USER  LEARNING_RATE  LATEST COMMIT
c9f380d     16 seconds ago  stopped  ben   0.01           d4fb0d3 (step 99)
a7cd781     9 seconds ago   stopped  ben   0.2            1f0865c (step 99)
```

:::note
Similar to how Git works, all this data is in your working directory. Replicate only lists experiments in the current working directory, so you'll only see experiments from the model you're working on.

If you want to poke around at the internal data, it is inside `.replicate/`. It's not something you'd do day to day, but there's no magic going on – everything's right there as files and plain JSON.
:::

As a reminder, this is a list of **experiments** which represents runs of the `train.py` script. Within experiments are **commits**, which are created every time you call `experiment.commit()` in your training script. The commit is the thing which actually contains the code and weights.

To list the commits within these experiments, you can use `replicate show`. Run this, replacing `c9f` with an experiment ID from your output of `replicate list`:

```shell-session
$ replicate show c9f380d
Experiment: c9f380d3530f5b5ba899827f137f25bcd3f81868f1416cf5c83f096ddee12530

Created:  Thu, 06 Aug 2020 11:55:54 PDT
Status:   stopped
Host:     107.133.144.125
User:     ben
Command:  train.py

Params
learning_rate:  0.01
num_epochs:     100

Commits
ID       STEP  CREATED      ACCURACY  LOSS
862e932  0     6 hours ago  0.33333   1.1836
dfdb97b  1     6 hours ago  0.33333   1.1173
e3650fe  2     6 hours ago  0.46667   1.0611
c811301  3     6 hours ago  0.63333   1.0138
...
71502b0  97    6 hours ago  1         0.12076
7cf044a  98    6 hours ago  1         0.11915
d4fb0d3  99    6 hours ago  1         0.1176
```

You can also use `replicate show` on commits to get all the information about it. Run this, replacing `d4f` with a commit ID from the experiment:

```shell-session
$ replicate show d4f
Commit: d4fb0d38114453337fb936a0c65cad63872f89e73c4e9161b666d59260848824

Created:  Thu, 06 Aug 2020 11:55:55 PDT
Step:     99

Experiment
ID:       c9f380d3530f5b5ba899827f137f25bcd3f81868f1416cf5c83f096ddee12530
Created:  Thu, 06 Aug 2020 11:55:54 PDT
Status:   stopped
Host:     107.133.144.125
User:     ben
Command:  train.py

Params
learning_rate:  0.01
num_epochs:     100

Labels
accuracy:  1
loss:      0.11759971082210541
```

:::note
Notice you can pass a prefix to `replicate show`, and it'll automatically find the experiment that starts with just those characters. Saves a few keystrokes.
:::

## Compare commits

Let's compare the last commits from the two experiments we ran. Run this, replacing `d4fb0d3` and `1f0865c` with the two commit IDs from the `LATEST COMMIT` column in `replicate ls`:

```shell-session
$ replicate diff d4fb0d3 1f0865c
Commit:                   d4fb0d3                   1f0865c
Experiment:               c9f380d                   a7cd781

Params
learning_rate:            0.01                      0.2

Labels
accuracy:                 1                         0.9666666388511658
loss:                     0.11759971082210541       0.056485891342163086
```

`replicate diff` works a bit like `git diff`, except in addition to the code, it compares all of the metadata that Replicate is aware of: params, metrics, dependencies, and so on.

:::note
`replicate diff` compares **commits**, because that is the thing that actually has all the results.

You can also pass an experiment ID, and it will pick the best or latest commit from that experiment.
:::

## Check out a commit

At some point you might want to get back to some point in the past. Maybe you've run a bunch of experiments in parallel, and you want to choose one that works best. Or, perhaps you've gone down a line of exploration and it's not working, so you want to get back to where you were a week ago.

The `replicate checkout` command will copy the files from a commit into your working directory. Run this, replacing `d4fb0d3` with a commit ID you passed to `replicate diff`:

```shell-session
$ replicate checkout d4fb0d3
═══╡ The directory "/Users/ben/p/tmp/iris-classifier" is not empty.
═══╡ This checkout may overwrite existing files. Make sure you've committed everything to Git so it's safe!

Do you want to continue? (y/N) y

═══╡ Checked out d4fb0d3 to "/Users/ben/p/tmp/iris-classifier"
```

The model file in your working directory is now the model saved in that commit:

```shell-session
$ ls -lh model.pth
-rw-r--r--  1 ben  staff   8.3K Aug  7 16:42 model.pth
```

This is useful for getting the trained model out of an experiment from the past, but **it also copies all of the code from that commit**. If you made a change to the code and didn't commit to Git, `replicate checkout` will allow you get back the exact code from an experiment.

**This means you don't have to remember to commit to Git when you're running experiments.** Just try a bunch of things, then when you've found something that works, use Replicate to get back to the exact code that produced those results and formally commit it to Git.

Neat, huh? Replicate is keeping track of everything in the background so you don't have to.

## The workflow so far

With these tools, let's recap what the workflow looks like:

- Add `experiment = replicate.init()` and `experiment.commit()` to your training code.
- Run several experiments by running the training script as usual, with changes to the hyperparameters or code.
- See the results of our experiments with `replicate ls` and `replicate show`.
- Compare the differences between experiments with `replicate diff`.
- Get the code from the best experiment with `replicate checkout`.
- Commit that code cleanly to Git.

You don't have to keep track of what you changed in your experiments, because Replicate does that automatically for you. You can also safely change things without committing to Git, because `replicate checkout` will always be able to get you back to the exact environment the experiment was run in.

## What's next

So far, everything we've been doing has been local on a single machine. But in practice, you probably want to train on a separate machine with GPUs, or perhaps multiple machines at the same time. Or, perhaps you want to share your experiments with other people in your team.

Take a look at [the guide to working with remote machines](working-with-remote-machines.md) to learn more about this.
