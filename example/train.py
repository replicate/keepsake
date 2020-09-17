import argparse
import replicate
from sklearn.datasets import load_iris
from sklearn.model_selection import train_test_split
from sklearn.utils import shuffle
import torch
from torch import nn
from torch.autograd import Variable


def train(learning_rate, num_epochs):
    # highlight-start
    # Create an "experiment". This represents a run of your training script.
    # It saves the training code at the given path and any hyperparameters.
    experiment = replicate.init(
        path=".",
        # highlight-start
        params={"learning_rate": learning_rate, "num_epochs": num_epochs},
    )
    # highlight-end

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

    torch.manual_seed(0)
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
        torch.save(model, "model.pth")
        # highlight-start
        # Create a checkpoint within the experiment.
        # This saves the metrics at that point, and makes a copy of the file
        # or directory given, which could weights and any other artifacts.
        experiment.checkpoint(
            path="model.pth",
            step=epoch,
            metrics={"loss": loss.item(), "accuracy": acc},
            primary_metric=("loss", "minimize"),
        )
        # highlight-end


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--learning_rate", type=float, default=0.01)
    parser.add_argument("--num_epochs", type=int, default=100)
    args = parser.parse_args()
    train(args.learning_rate, args.num_epochs)
