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

#### `replicate.keras_callback.ReplicateCallback(filepath="model.hdf5", params=None, primary_metric=None, save_freq="epoch")`

_Note: This is an experimental feature, and the API may change in the future._

Replicate provies a convenient callback when you're working with Keras. `ReplicateCallback` behaves like Tensorflow's [ModelCheckpoint callback](https://www.tensorflow.org/api_docs/python/tf/keras/callbacks/ModelCheckpoint), but in addition to exporting a model at the end of each epoch, it also:
- runs `replicate.init()` to initialize an experiment at the start of training, and
- runs `experiment.checkpoint()` after saving the model at the end of the epoch (or every _n_ batches if `save_freq` is an integer). This call also saves checkpoint metadata, that it gets from the [`logs` dictionary](https://www.tensorflow.org/guide/keras/custom_callback#a_basic_example) that is passed to the callback's `on_epoch_end` method.

This class takes the following arguments:
- `filepath`: The path where the exported model is saved. This path is also saved by `experiment.checkpoint()` at the end of each epoch.
- `params`: A dictionary of hyperparameters that will be recorded to the experiment at the start of training.
- `primary_metric`: A _pair_ in the format `(metric_name, goal)`, where `goal` is either `minimize`, or `maximize`.
- `save_freq`:`"epoch"` or integer. When using `"epoch"`, the callback saves the model after each epoch. When using integer, the callback saves the model at end of this many batches.

Example:

```python
dense_size = 784
learning_rate = 0.01

# from https://www.tensorflow.org/guide/keras/custom_callback
model = keras.Sequential()
model.add(keras.layers.Dense(1, input_dim=dense_size))
model.compile(
    optimizer=keras.optimizers.RMSprop(learning_rate=learning_rate),
    loss="mean_squared_error",
    metrics=["mean_absolute_error"],
)

# Load example MNIST data and pre-process it
(x_train, y_train), (x_test, y_test) = tf.keras.datasets.mnist.load_data()
x_train = x_train.reshape(-1, 784).astype("float32") / 255.0
x_test = x_test.reshape(-1, 784).astype("float32") / 255.0

model.fit(
    x_train[:1000],
    y_train[:1000],
    batch_size=128,
    epochs=20,
    validation_split=0.5,
    callbacks=[
        MyLogger(),
        ReplicateCallback(
            params={"dense_size": dense_size, "learning_rate": learning_rate,},
            primary_metric=("mean_absolute_error", "minimize"),
            save_freq=2,
        ),
    ],
)
```
