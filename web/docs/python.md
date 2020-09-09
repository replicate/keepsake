---
id: python
title: Python library reference
---

import CodeBlock from "@theme/CodeBlock";
import config from '../docusaurus.config.js';

The Replicate Python library is used to create experiments and checkpoints in your training script.

This page is a comprehensive reference for it. To get an introduction on how to use it, see [the tutorial](tutorial.md).

## Install

You can install the library by adding it to your `requirements.txt` file:

<CodeBlock className="txt">
{`replicate==`+config.customFields.version+`
`}</CodeBlock>
 

It's a good idea to put it in a file and commit it to Git so other people who use your code will have it installed automatically.

You can also use pip if you are installing it locally:

```
pip install replicate
```

## API

#### `replicate.init(params)`

Create and return an experiment.

It takes one argument: `params`, a dictionary of hyperparameters to record along with the experiment.

The entire project directory will be saved to storage to save a copy of the code. The project directory is determined by the directory that contains `replicate.yaml`. If no `replicate.yaml` is found in any parent directories, the current working directory will be used.

To determine where data will be stored, this method will read `replicate.yaml` file in the working directory and use the `storage` option. [Learn more in the reference documentation.](replicate-yaml.md)

For example:

```python
>>> import replicate
>>> experiment = replicate.init(params={"learning_rate" 0.01})
```

#### `experiment.checkpoint(path, metrics, primary_metric=None, step=None)`

Create a checkpoint within an experiment.

It takes these arguments:

- `path`: A path to a file or directory that will be uploaded to storage, relative to the working directory. This can be used to save weights, Tensorboard logs, and other artifacts produced during the training process. If `path` is `None`, no data will be saved.
- `metrics`: A dictionary of metrics to record along with the checkpoint.
- `primary_metric` (optional): A tuple `(name, goal)` to define one of the metrics as a primary metric to optimize. Goal can either be `minimize` or `maximize`.
- `step` (optional): the iteration number of this checkpoint, such as epoch number. This is displayed in `replicate ls` and various other places.

Any keyword arguments passed to the function will also be recorded.

You can add more information about these keyword arguments in `replicate.yaml` to define which ones are metrics that determine the performance of your model, and what success is for that metric. [See the `replicate.yaml` documentation for more information.](replicate-yaml.md#metrics).

For example:

```python
>>> experiment.checkpoint(
...   path="weights/",
...   step=5,
...   metrics={"train_loss": 0.425, "train_accuracy": 0.749},
...   primary_metric=("train_accuracy", "maximize"),
... )
```
