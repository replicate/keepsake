from copy import deepcopy
from typing import Optional, Dict, Tuple, Any
from pathlib import Path

import keepsake
from pytorch_lightning.callbacks.base import Callback


class KeepsakeCallback(Callback):
    """
    Experimental class that wraps keepsake.init() and
    experiment.checkpoint() in a PyTorch Lightning callback.
    API is mostly consistent with keras_callback.KeepsakeCallback.

    This integration is subject to change, we'll probably
    try to get it into Pytorch Lightning itself as a logger.
    See https://github.com/replicate/keepsake/issues/432

    The KeepsakeCallback instantiates a new Keepsake experiment with
    keepsake.init on_pretrain_routine_start, using the parameters
    that are passed to the KeepsakeCallback constructor.

    Checkpoints are saved on_validation_end if validation is defined,
    otherwise on_epoch_end.

    For an example of how to use this callback, see the guide at
    https://keepsake.ai/docs/guides/pytorch-lightning-integration
    """

    def __init__(
        self,
        filepath="model.pth",
        experiment_path=".",
        params: Optional[Dict[str, Any]] = None,
        primary_metric: Optional[Tuple[str, str]] = None,
        period: Optional[int] = 1,
        save_weights_only: Optional[bool] = False,
    ):
        """
        Create a new Pytorch Lightning Keepsake callback.

        Parameters
        ----------
        filepath : str, default "model.pth"
            The name of the model artifact to save. To disable
            saving model artifacts, set filepath=None
        params : dict, default None
            Hyperparameters and other metadata to save in the
            experiment.
        primary_metric : tuple (name, goal), default None
            Name and optimization goal of the primary metric.
            Goal must be either "minimize" or "maximize".
            The metric must be logged some time during the
            training process.
        save_weights_only : bool, default False
            if True, then only the model’s weights will be saved,
            else the full model is saved.
        """

        super().__init__()
        self.filepath = Path(filepath).resolve()
        self.experiment_path = Path(experiment_path).resolve()
        self.params = params
        self.primary_metric = primary_metric
        self.period = period
        self.save_weights_only = save_weights_only
        self.last_global_step_saved = -1

    def on_pretrain_routine_start(self, trainer, pl_module):
        self.experiment = keepsake.init(
            path=str(self.experiment_path),
            params=self.params,
        )

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
            trainer.save_checkpoint(self.filepath.name, self.save_weights_only)

        self.last_global_step_saved = global_step

        metrics = deepcopy(trainer.logger_connector.logged_metrics)
        metrics.update(trainer.logger_connector.callback_metrics)
        metrics.update(trainer.logger_connector.progress_bar_metrics)
        metrics.update({"global_step": trainer.global_step})

        self.experiment.checkpoint(
            path=self.filepath.name,
            step=trainer.current_epoch,
            metrics=metrics,
            primary_metric=self.primary_metric,
        )
