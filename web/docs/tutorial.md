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
mkdir replicate-getting-started
cd replicate-getting-started
```

Then, copy and paste this code into `train.py`:

```python title="train.py" {29,33-39}
import argparse
from fastai import tabular
import pandas
import replicate
from sklearn.model_selection import train_test_split


def train(learning_rate):
    df = pandas.read_csv(
        "bezdekIris.data",
        names=["sepal_length", "sepal_width", "petal_length", "petal_width", "class"],
    )
    dep_var = "class"
    cont_names = ["sepal_length", "sepal_width", "petal_length", "petal_width"]
    procs = [tabular.FillMissing, tabular.Categorify, tabular.Normalize]

    train_df, test_df = train_test_split(df, test_size=0.2, random_state=1)
    test = tabular.TabularList.from_df(test_df, cont_names=cont_names)
    data = (
        tabular.TabularList.from_df(train_df, cont_names=cont_names, procs=procs)
        .split_by_rand_pct(valid_pct=0.2, seed=1)
        .label_from_df(cols=dep_var)
        .add_test(test)
        .databunch()
    )

    learn = tabular.tabular_learner(data, layers=[200, 100], metrics=tabular.accuracy)

    experiment = replicate.init(params={"learning_rate": learning_rate})

    class ReplicateCallback(tabular.Callback):
        def on_epoch_end(self, epoch, last_loss, last_metrics, **kwargs):
            experiment.commit(
                metrics={
                    "train_loss": float(last_loss.item()),
                    "val_loss": last_metrics[0].astype(float),
                    "accuracy": float(last_metrics[1].item()),
                }
            )

    learn.fit(50, lr=learning_rate, callbacks=[ReplicateCallback()])


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--learning_rate", type=float, default=0.01)
    train(**vars(parser.parse_args()))
```

Notice there are two statements in this training code that call Replicate, highlighted.

The first is `replicate.init()`. This creates an **experiment**, which represents the training run. This is called just once in your training script at the start so you can pass your hyperparameters.

The second is `experiment.commit()`. This creates a **commit**, which saves the exact state of the filesystem at that point (code, weights, Tensorboard logs, etc), along with some metrics you pass to the function. An experiment will typically contain multiple commits, and they're typically done on every epoch when you might save your model file.

## Download training data and dependencies

Before we start training, we need to download training data and install the Python dependencies.

Download the training data:

```
wget https://archive.ics.uci.edu/ml/machine-learning-databases/iris/bezdekIris.data
```

Create `requirements.txt` to define our requirements:

```txt title="requirements.txt"
pandas==1.0.5
fastai==1.0.61
scikit-learn==0.23.1
https://storage.googleapis.com/bfirsh-dev/replicate-python-2/replicate-0.0.1.tar.gz
```

Then, install the Python requirements inside a Virtualenv:

```
virtualenv venv
. venv/bin/activate
pip install -r requirements.txt
```

## Start training

We're now going to train this model a couple of times with different parameters to see what we can do with Replicate.

First, train it with default parameters:

```
$ python train.py
epoch     train_loss  valid_loss  accuracy  time
0         1.102223    0.367693    0.916667  00:00
1         0.688286    0.281458    0.916667  00:00
2         0.508243    0.329499    0.916667  00:00
3         0.420214    0.410820    0.875000  00:00
4         0.355645    0.424739    0.875000  00:00
5         0.312976    0.417650    0.875000  00:00
...
```

Next, run the training with a different learning rate:

```
$ python train.py --learning_rate=0.05
epoch     train_loss  valid_loss  accuracy  time
0         1.188458    0.173192    0.916667  00:00
1         0.795049    0.584164    0.750000  00:00
4         0.388231    0.926992    0.875000  00:00
5         0.324020    0.919221    0.875000  00:00
...
```

## List and view experiments

## Compare experiments

## Check out experiments

## What's next

So far, everything we've been doing has been local on a single machine. But in practice, you probably want to train on a separate machine with GPUs, or perhaps multiple machines at the same time. Or, perhaps you want to share your experiments with other people in your team.

Take a look at [the guide to working with remote machines](working-with-remote-machines.md) to learn more about this.
