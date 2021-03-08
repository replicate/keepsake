import datetime

import matplotlib.pyplot as plt
from keepsake.checkpoint import Checkpoint, CheckpointList
from keepsake.experiment import Experiment, ExperimentList
from keepsake.project import Project, init


def test_num_plots(temp_workdir):
    with open("keepsake.yaml", "w") as f:
        f.write("repository: file://.keepsake/")

    experiment = init(path=".", params={"learning_rate": 0.1, "num_epochs": 1},)

    experiment.checkpoint(
        path=".",
        step=1,
        metrics={"loss": 1.1836304664611816, "accuracy": 0.3333333432674408},
        primary_metric=("loss", "minimize"),
    )
    experiment.checkpoint(
        path=".",
        step=2,
        metrics={"loss": 1.1836304662222222, "accuracy": 0.4333333432674408},
        primary_metric=("loss", "minimize"),
    )

    experiment_list = ExperimentList([experiment])
    num_plots = 30
    for rep in range(num_plots):
        experiment_list.plot()
    assert len(plt.get_fignums()) == 1
    experiment_list.plot(metric="accuracy")
    assert len(plt.get_fignums()) == 2
