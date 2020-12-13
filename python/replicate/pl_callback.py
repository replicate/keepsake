from copy import deepcopy
from typing import Optional, Dict

import replicate
from pytorch_lightning.callbacks.base import Callback


class ReplicateCallback(Callback):
    def __init__(
        self,
        params: Dict,
        monitor: str,
        period: Optional[int] = 1,
        save_weights_only: Optional[bool] = False,
        mode: Optional[str] = 'min'
    ):
        super().__init__()
        self.params = params
        self.monitor = monitor
        self.period = period
        self.save_weights_only = save_weights_only
        self.mode = 'minimize' if mode == 'min' else 'maximize'

    def on_pretrain_routine_start(self, trainer, pl_module):
        self.experiment = replicate.init(path='.', params=self.params)

    def on_validation_end(self, trainer, pl_module):
        if trainer.current_epoch % self.period == 0 and trainer.current_epoch:
            self._save_model('model.pth', trainer, pl_module)

    def _save_model(self, filepath: str, trainer, pl_module):
        trainer.save_checkpoint(filepath, self.save_weights_only)

        metrics = deepcopy(trainer.logger_connector.logged_metrics)
        metrics.update(trainer.logger_connector.callback_metrics)
        metrics.update(trainer.logger_connector.progress_bar_metrics)
        metrics.update({
            "step": trainer.global_step,
            "epoch": trainer.current_epoch
        })

        self.experiment.checkpoint(
            path=filepath,
            step=trainer.current_epoch,
            metrics=metrics,
            primary_metric=(self.monitor, self.mode)
        )