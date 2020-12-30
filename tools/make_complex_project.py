import argparse
import os
import tempfile
from sklearn.datasets import load_iris
from sklearn.model_selection import train_test_split
from sklearn.utils import shuffle
import torch
from torch import nn
from torch.autograd import Variable


from replicate.project import Project
from replicate.repository import repository_for_url


def train(project, learning_rate, num_epochs, hidden_layer_size=15, **kwargs):
    params = {
        "learning_rate": learning_rate,
        "num_epochs": num_epochs,
        "hidden_layer_size": hidden_layer_size,
    }
    params.update(kwargs)
    experiment = project.experiments.create(path=None, params=params)

    print("Downloading data set...")
    iris = load_iris()
    train_features, val_features, train_labels, val_labels = train_test_split(
        iris.data,
        iris.target,
        train_size=0.8,
        test_size=0.2,
        random_state=0,
        stratify=iris.target,
    )
    train_features = torch.FloatTensor(train_features)
    val_features = torch.FloatTensor(val_features)
    train_labels = torch.LongTensor(train_labels)
    val_labels = torch.LongTensor(val_labels)

    model = nn.Sequential(nn.Linear(4, 15), nn.ReLU(), nn.Linear(15, 3),)
    optimizer = torch.optim.Adam(model.parameters(), lr=learning_rate)
    criterion = nn.CrossEntropyLoss()

    for epoch in range(num_epochs):
        model.train()
        optimizer.zero_grad()
        outputs = model(train_features)
        loss = criterion(outputs, train_labels)
        loss.backward()
        optimizer.step()

        with torch.no_grad():
            model.eval()
            output = model(val_features)
            acc = (output.argmax(1) == val_labels).float().sum() / len(val_labels)

        print(
            "Epoch {}, train loss: {:.3f}, validation accuracy: {:.3f}".format(
                epoch, loss.item(), acc
            )
        )
        experiment.checkpoint(
            path=None,
            step=epoch,
            metrics={"loss": loss.item(), "accuracy": acc},
            primary_metric=("loss", "minimize"),
        )

    experiment.stop()


parser = argparse.ArgumentParser(
    description="Create a project with a bunch of realistic-ish data to test `replicate ls` output and things"
)
parser.add_argument("repository")
args = parser.parse_args()

with tempfile.TemporaryDirectory() as project_dir:
    print("Creating project...")
    project = Project(directory=".", repository=args.repository)
    train(project, learning_rate=0.01, num_epochs=10)
    train(project, learning_rate=0.05, num_epochs=10)
    train(project, learning_rate=0.01, num_epochs=100)
    train(project, learning_rate=0.05, num_epochs=100)
    train(project, learning_rate=0.001, num_epochs=100)
    train(project, learning_rate=0.1, num_epochs=100)
    train(project, learning_rate=0.01, num_epochs=100, hidden_layer_size=30)
    train(project, learning_rate=0.01, num_epochs=100, hidden_layer_size=10)
    train(project, learning_rate=0.01, num_epochs=50, model="bartnet")
    train(project, learning_rate=0.01, num_epochs=50, dropout_rate=0.5, model="homnet")
    train(project, learning_rate=0.01, num_epochs=50, dropout_rate=0.3, model="homnet")
    train(
        project,
        learning_rate=0.01,
        num_epochs=50,
        dropout_rate=0.3,
        model="moenet",
        decay_rate=0.3,
    )
    train(
        project,
        learning_rate=0.01,
        num_epochs=50,
        dropout_rate=0.3,
        model="moenet",
        decay_rate=0.5,
    )
    train(
        project,
        learning_rate=0.01,
        num_epochs=50,
        dropout_rate=0.3,
        model="moenet",
        decay_rate=0.1,
    )
    train(
        project,
        learning_rate=0.01,
        num_epochs=50,
        dropout_rate=0.3,
        model="marnet",
        decay_rate=0.1,
    )
    train(
        project,
        learning_rate=0.05,
        num_epochs=50,
        dropout_rate=0.3,
        model="marnet",
        decay_rate=0.1,
    )
    train(
        project,
        learning_rate=0.05,
        num_epochs=50,
        dropout_rate=0.3,
        model="marnet",
        decay_rate=0.1,
        description="It was the best of times, it was the worst of times, it was the age of wisdom, it was the age of foolishness, it was the epoch of belief, it was the epoch of incredulity, it was the season of light, it was the season of darkness, it was the spring of hope, it was the winter of despair.",
    )
    train(
        project,
        learning_rate=0.05,
        num_epochs=50,
        dropout_rate=0.3,
        model="marnet",
        decay_rate=0.1,
        description="It was the best of times, it was the blurst of times",
    )

    # print("Uploading to repository...")
    # repository = repository_for_url(args.repository)
    # repository.put_path(os.path.join(project_dir, ".replicate/"), "")
