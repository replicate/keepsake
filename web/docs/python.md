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

Any keyword arguments will be recorded as hyperparameters. The experiment will be saved to storage, with the hyperparameters you pass and various other automatically generated metadata.

To determine the storage location, this method will look for `replicate.yaml` file in the working directory. [Learn more in the reference documentation.](replicate-yaml.md)

For example:

```python
>>> import replicate
>>> experiment = replicate.init(learning_rate=0.01)
```

#### `experiment.commit(**labels)`

Create a commit within an experiment.

When this is called, Replicate takes a copy of the working directory and uploads it to storage. Any keyword arguments passed to the function will also be recorded.

You can add more information about these keyword arguments in `replicate.yaml` to define which ones are metrics that determine the performance of your model, and what success is for that metric. [See the `replicate.yaml` documentation for more information.](replicate-yaml.md#metrics).

For example:

```python
>>> experiment.commit(step=5, train_loss=0.425, train_accuracy=0.749)
```
