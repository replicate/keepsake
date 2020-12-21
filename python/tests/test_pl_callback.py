import json
from glob import glob
import os
import torch
from torch.nn import functional as F
from torch import nn
from pytorch_lightning.core.lightning import LightningModule
from pytorch_lightning import Trainer
from torch.optim import Adam
from torch.utils.data import DataLoader, random_split, Subset
from torch.utils.data.dataset import Dataset
from torchvision.datasets import MNIST
from torchvision import transforms
import pytest

from replicate.pl_callback import ReplicateCallback


class ModelNoValidation(LightningModule):
    def __init__(self):
        super().__init__()

        self.layer_1 = torch.nn.Linear(28 * 28, 128)
        self.layer_2 = torch.nn.Linear(128, 10)
        self.batch_size = 8

    def forward(self, x):
        batch_size = x.size()[0]

        x = x.view(batch_size, -1)
        x = self.layer_1(x)
        x = F.relu(x)
        x = self.layer_2(x)
        x = F.log_softmax(x, dim=1)
        return x

    def prepare_data(self):
        # download only
        MNIST(
            "/tmp/replicate-test-mnist",
            train=True,
            download=True,
            transform=transforms.ToTensor(),
        )

    def setup(self, stage):
        # transform
        transform = transforms.Compose([transforms.ToTensor()])
        mnist_train = MNIST(
            "/tmp/replicate-test-mnist", train=True, download=False, transform=transform
        )
        mnist_train = Subset(mnist_train, range(100))

        # train/val split
        mnist_train, mnist_val = random_split(mnist_train, [80, 20])

        # assign to use in dataloaders
        self.train_dataset = mnist_train
        self.val_dataset = mnist_val

    def train_dataloader(self):
        return DataLoader(self.train_dataset, batch_size=self.batch_size)  # type: ignore

    def training_step(self, batch, batch_idx):
        x, y = batch
        logits = self(x)
        loss = F.nll_loss(logits, y)
        self.log("train_loss", x, on_step=True, on_epoch=True, logger=False)
        return loss

    def configure_optimizers(self):
        return Adam(self.parameters(), lr=1e-3)


class ModelWithValidation(ModelNoValidation):
    def val_dataloader(self):
        return DataLoader(self.val_dataset, batch_size=self.batch_size)  # type: ignore

    def validation_step(self, batch, batch_idx):
        x, y = batch
        logits = self(x)
        loss = F.nll_loss(logits, y)
        self.log("val_loss", x, on_step=True, on_epoch=True, logger=False)
        return loss


def test_pl_callback_no_validation(temp_workdir):
    with open("replicate.yaml", "w") as f:
        f.write("repository: file://.replicate/")

    dense_size = 784
    learning_rate = 0.1

    model = ModelNoValidation()
    trainer = Trainer(
        checkpoint_callback=False,
        callbacks=[
            ReplicateCallback(
                params={"dense_size": dense_size, "learning_rate": learning_rate,},
                primary_metric=("train_loss", "minimize"),
            )
        ],
        max_epochs=5,
    )

    trainer.fit(model)

    exp_meta_paths = glob(".replicate/metadata/experiments/*.json")
    assert len(exp_meta_paths) == 1
    with open(exp_meta_paths[0]) as f:
        exp_meta = json.load(f)
    assert exp_meta["params"]["dense_size"] == 784
    assert exp_meta["params"]["learning_rate"] == 0.1
    assert len(exp_meta["checkpoints"]) == 5
    chkp_meta = exp_meta["checkpoints"][0]

    assert chkp_meta["path"] == "model.pth"
    assert chkp_meta["primary_metric"] == {
        "name": "train_loss",
        "goal": "minimize",
    }
    assert set(chkp_meta["metrics"].keys()) == set(
        ["train_loss", "global_step", "epoch",]
    )
    assert os.path.exists(".replicate/checkpoints/" + chkp_meta["id"] + ".tar.gz")


def test_pl_callback_with_validation(temp_workdir):
    with open("replicate.yaml", "w") as f:
        f.write("repository: file://.replicate/")

    dense_size = 784
    learning_rate = 0.1

    model = ModelWithValidation()
    trainer = Trainer(
        checkpoint_callback=False,
        callbacks=[
            ReplicateCallback(
                params={"dense_size": dense_size, "learning_rate": learning_rate,},
                primary_metric=("val_loss", "minimize"),
            )
        ],
        max_epochs=5,
    )

    trainer.fit(model)

    exp_meta_paths = glob(".replicate/metadata/experiments/*.json")
    assert len(exp_meta_paths) == 1
    with open(exp_meta_paths[0]) as f:
        exp_meta = json.load(f)
    assert exp_meta["params"]["dense_size"] == 784
    assert exp_meta["params"]["learning_rate"] == 0.1
    assert len(exp_meta["checkpoints"]) == 5
    chkp_meta = exp_meta["checkpoints"][0]

    assert chkp_meta["path"] == "model.pth"
    assert chkp_meta["primary_metric"] == {
        "name": "val_loss",
        "goal": "minimize",
    }
    assert set(chkp_meta["metrics"].keys()) == set(
        ["val_loss", "global_step", "epoch",]
    )
    assert os.path.exists(".replicate/checkpoints/" + chkp_meta["id"] + ".tar.gz")
