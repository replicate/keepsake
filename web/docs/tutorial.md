---
id: tutorial
title: Tutorial
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

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

    curl -o /usr/local/bin/replicate https://storage.googleapis.com/replicate-public/cli/latest/darwin/amd64/replicate
    chmod +x /usr/local/bin/replicate

</TabItem>
<TabItem value="linux">
Run the following commands in a terminal:

    sudo curl -o /usr/local/bin/replicate https://storage.googleapis.com/replicate-public/cli/latest/linux/amd64/replicate
    sudo chmod +x /usr/local/bin/replicate

</TabItem>
</Tabs>

## Write model

We're going to make a model that classifies Iris plants, trained on the [Iris dataset](https://archive.ics.uci.edu/ml/datasets/iris). It's an intentionally simple model that trains really fast, just so we can show you how Replicate works.

First, let's make a directory to work in:

```
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
    train(**vars(parser.parse_args()))
```

Notice there are two statements in this training code that call Replicate, highlighted.

The first is `replicate.init()`. This creates an **experiment**, which represents the training run. This is called just once in your training script at the start so you can pass your hyperparameters.

The second is `experiment.commit()`. This creates a **commit**, which saves the exact state of the filesystem at that point (code, weights, Tensorboard logs, etc), along with some metrics you pass to the function. An experiment will typically contain multiple commits, and they're typically done on every epoch when you might save your model file.

## Install dependencies

Before we start training, we need to install the Python dependencies.

Create `requirements.txt` to define our requirements:

```txt title="requirements.txt"
https://storage.googleapis.com/replicate-python-dev/replicate-0.0.9.tar.gz
scikit-learn==0.23.1
torch==1.5.1
```

Then, install the Python requirements inside a Virtualenv:

```
virtualenv venv
. venv/bin/activate
pip install -r requirements.txt
```

## Train the model

We're now going to train this model a couple of times with different parameters to see what we can do with Replicate.

First, train it with default parameters:

```
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

```
$ python train.py --learning_rate=0.2
Epoch 0, train loss: 1.184, validation accuracy: 0.333
Epoch 1, train loss: 1.161, validation accuracy: 0.633
Epoch 2, train loss: 1.124, validation accuracy: 0.667
...
Epoch 97, train loss: 0.057, validation accuracy: 0.967
Epoch 98, train loss: 0.057, validation accuracy: 0.967
Epoch 99, train loss: 0.056, validation accuracy: 0.967
```

## List and view experiments

By default, the calls to the `replicate` Python library have saved your experiments to your local disk. You can use `replicate list` to list them:

    $ replicate list
    EXPERIMENT  STARTED         STATUS   USER  LEARNING_RATE  LATEST COMMIT
    c9f380d     16 seconds ago  stopped  ben   0.01           d4fb0d3 (step 99)
    a7cd781     9 seconds ago   stopped  ben   0.2            1f0865c (step 99)

As a reminder, this is a list of **experiments** which represents runs of the `train.py` script. Within experiments are **commits**, which are created every time you call `experiment.commit()` in your training script. The commit is the thing which actually contains the code and weights.

To list the commits within these experiments, you can use `replicate show`:

    $ replicate show c9f
    Experiment: c9f380d3530f5b5ba899827f137f25bcd3f81868f1416cf5c83f096ddee12530

    Created:  Thu, 06 Aug 2020 11:55:54 PDT
    Status:   stopped
    Host:     107.133.144.125
    User:     ben

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

Notice how you can pass a prefix to `replicate show`, and it'll automatically find the experiment that starts with just those characters. Saves a few keystrokes.

You can also use `replicate show` on commits to get all the information about it:

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

    Params
    learning_rate:  0.01
    num_epochs:     100

    Labels
    accuracy:  1
    loss:      0.11759971082210541

Similar to how Git works, all this data is in the current directory, so you'll only see the model you're working on. (If you want to poke around, it's inside `.replicate/` in your working directory.)

## Compare experiments

## Check out experiments

## What's next

So far, everything we've been doing has been local on a single machine. But in practice, you probably want to train on a separate machine with GPUs, or perhaps multiple machines at the same time. Or, perhaps you want to share your experiments with other people in your team.

Take a look at [the guide to working with remote machines](working-with-remote-machines.md) to learn more about this.
