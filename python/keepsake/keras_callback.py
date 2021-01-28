from typing import Optional, Dict, Any, Tuple

try:
    from tensorflow.keras.callbacks import ModelCheckpoint  # type: ignore
except ImportError:
    from keras.callbacks import ModelCheckpoint  # type: ignore

from . import init
from .experiment import Experiment
from .project import Project


class KeepsakeCallback(ModelCheckpoint):
    """
    Experimental class that wraps keepsake.init() and
    experiment.checkpoint() in a Keras callback.
    """

    def __init__(
        self,
        filepath="model.hdf5",
        params: Optional[Dict[str, Any]] = None,
        primary_metric: Optional[Tuple[str, str]] = None,
        save_freq: str = "epoch",
        save_weights_only: bool = False,
    ):
        self.init_params = params
        self.primary_metric = primary_metric
        self.experiment: Experiment
        self.step = 0

        super().__init__(
            filepath=filepath,
            verbose=0,
            save_best_only=False,
            save_weights_only=save_weights_only,
            save_freq=save_freq,
        )

    def on_train_begin(self, logs=None):
        self.experiment = init(path=".", params=self.init_params)
        super().on_train_begin(logs)

    # FIXME(bfirsh): probably a bad idea to override an internal method -- this theoretically might break at some point?
    def _save_model(self, epoch, logs):
        logs = logs or {}

        filepath = None
        if self.filepath is not None:
            if not (
                isinstance(self.save_freq, int)
                or self.epochs_since_last_save >= self.period
            ):
                return

            self.epochs_since_last_save = 0

            filepath = self._get_file_path(epoch, logs)

            try:
                if self.save_weights_only:
                    self.model.save_weights(filepath, overwrite=True)
                else:
                    self.model.save(filepath, overwrite=True)

                self._maybe_remove_file()
            except IOError as e:
                # `e.errno` appears to be `None` so checking the content
                # of `e.args[0]`. This error check is from
                # keras.callbacks.ModelCheckpoint
                if "is a directory" in e.args[0]:
                    raise IOError(
                        "Please specify a non-directory filepath for "
                        "ModelCheckpoint. Filepath used is an existing "
                        "directory: {}".format(filepath)
                    )

        self.experiment.checkpoint(
            filepath,
            step=self.step,
            metrics=logs,
            primary_metric=self.primary_metric,
            # TODO(bfirsh): make this output information in a way that is compatible with keras's progress bar
            quiet=True,
        )

        # if save_freq is an integer, step = batch_number * save_freq
        # if save_freq is epoch, step = epoch_number
        self.step += 1
