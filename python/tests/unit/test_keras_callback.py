import json
from glob import glob
import tensorflow as tf  # type: ignore
from tensorflow import keras  # type: ignore
import pytest

from replicate.keras_callback import ReplicateCallback

from .common import temp_workdir


def test_keras_callback(temp_workdir):
    dense_size = 784
    learning_rate = 0.1

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

    # Limit the data to 1000 samples
    x_train = x_train[:1000]
    y_train = y_train[:1000]
    x_test = x_test[:1000]
    y_test = y_test[:1000]

    model.fit(
        x_train,
        y_train,
        batch_size=128,
        epochs=5,
        verbose=0,
        validation_split=0.5,
        callbacks=[
            ReplicateCallback(
                params={"dense_size": dense_size, "learning_rate": learning_rate,},
                primary_metric=("mean_absolute_error", "minimize"),
            )
        ],
    )

    exp_meta_paths = glob(".replicate/storage/metadata/experiments/*.json")
    assert len(exp_meta_paths) == 1
    with open(exp_meta_paths[0]) as f:
        exp_meta = json.load(f)
    assert exp_meta["params"]["dense_size"] == 784
    assert exp_meta["params"]["learning_rate"] == 0.1

    chkp_meta_paths = glob(".replicate/storage/metadata/checkpoints/*.json")
    assert len(chkp_meta_paths) == 5
    with open(chkp_meta_paths[0]) as f:
        chkp_meta = json.load(f)
    assert chkp_meta["path"] == "model.hdf5"
    assert chkp_meta["primary_metric"] == {
        "name": "mean_absolute_error",
        "goal": "minimize",
    }
    assert set(chkp_meta["metrics"].keys()) == set(
        ["mean_absolute_error", "loss", "val_mean_absolute_error", "val_loss"]
    )