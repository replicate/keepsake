---
id: python
title: Python library reference
---

import CodeBlock from "@theme/CodeBlock";
import config from '../docusaurus.config.js';

The Replicate Python library is used to create experiments and commits in your training script.

This page is a comprehensive reference for it. To get an introduction on how to use it, see [the tutorial](tutorial.md).

## Install

You can install the library by adding it to your `requirements.txt` file:

<CodeBlock className="txt">
{`replicate==`+config.customFields.version+`
`}</CodeBlock>
 

We recommend putting it in a file and committing it to Git so other people who use your code will have it installed automatically.

You can also use pip if you are installing it locally:

```
pip install replicate
```

## API

#### `replicate.init(**params)`

Create and return an experiment.

The entire working directory will be saved to storage to save a copy of the code. Any keyword arguments will be recorded as hyperparameters.

To determine the storage location, this method will read `replicate.yaml` file in the working directory and use the `storage` option. [Learn more in the reference documentation.](replicate-yaml.md)

For example:

```python
>>> import replicate
>>> experiment = replicate.init(learning_rate=0.01)
```

#### `experiment.commit(path, **labels)`

Create a commit within an experiment.

The given `path`, relative to the working directory, will be uploaded to storage. It can be either a directory or a single file. This can be used to save weights, Tensorboard logs, and other artifacts produced during the training process. If `path` is `None`, no data will be saved.

Any keyword arguments passed to the function will also be recorded.

You can add more information about these keyword arguments in `replicate.yaml` to define which ones are metrics that determine the performance of your model, and what success is for that metric. [See the `replicate.yaml` documentation for more information.](replicate-yaml.md#metrics).

For example:

```python
>>> experiment.commit(path="weights/", step=5, train_loss=0.425, train_accuracy=0.749)
```
