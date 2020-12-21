from copy import deepcopy
from typing import Optional, Dict, Tuple

import replicate
from pytorch_lightning.callbacks.base import Callback


class ReplicateCallback(Callback):
    """
    Experimental class that wraps replicate.init() and
    experiment.checkpoint() in a PyTorch Lightning callback.
    API is mostly consistent with keras_callback.ReplicateCallback.

    Checkpoints are saved on_validation_end if validation is defined,
    otherwise on_epoch_end.
    """

    def __init__(
        self,
        filepath="model.pth",
        params: Optional[Dict] = None,
        primary_metric: Optional[Tuple[str, str]] = None,
        period: Optional[int] = 1,
        save_weights_only: Optional[bool] = False,
    ):
        super().__init__()
        self.filepath = filepath
        self.params = params
        self.primary_metric = primary_metric
        self.period = period
        self.save_weights_only = save_weights_only
        self.last_global_step_saved = -1

    def on_pretrain_routine_start(self, trainer, pl_module):
        self.experiment = replicate.init(path=".", params=self.params)

    def on_epoch_end(self, trainer, pl_module):
        self._save_model(trainer, pl_module)

    def on_validation_end(self, trainer, pl_module):
        self._save_model(trainer, pl_module)

    def _save_model(self, trainer, pl_module):
        epoch = trainer.current_epoch
        global_step = trainer.global_step

        if (
            # no models are saved
            self.period < 1
            # skip epoch
            or (epoch + 1) % self.period
            # don't save anything during sanity check
            or trainer.running_sanity_check
            # already saved at the last step
            or self.last_global_step_saved == global_step
        ):
            return

        if self.filepath != None:
            trainer.save_checkpoint(self.filepath, self.save_weights_only)

        self.last_global_step_saved = global_step

        metrics = deepcopy(trainer.logger_connector.logged_metrics)
        metrics.update(trainer.logger_connector.callback_metrics)
        metrics.update(trainer.logger_connector.progress_bar_metrics)
        metrics.update(
            {"global_step": trainer.global_step,}
        )

        self.experiment.checkpoint(
            path=self.filepath,
            step=trainer.current_epoch,
            metrics=metrics,
            primary_metric=self.primary_metric,
        )
