# Replicate

Version control for machine learning.

Just add two lines to your training code:

```python
import torch
import replicate

def train():
    # Save training code and hyperparameters
    experiment = replicate.init(path=".", params={...})
    model = Model()

    for epoch in range(num_epochs):
        # ...

        torch.save(model, "model.pth")
        # Save model weights and metrics
        experiment.checkpoint(path="model.pth", metrics={...})
```

Then Replicate will start tracking everything: code, hyperparameters, training data, weights, metrics, Python dependencies, and so on.

- **It's open source:** It won't stop working if a startup goes out of business.
- **You're in control of your data:** All the data is stored on your own Amazon S3 or Google Cloud Storage as plain old files. There's no server to run.
- **It works with everything:** Tensorflow, PyTorch, scikit-learn, XGBoost, you name it. It's just saving files and dictionaries – export however you want.

## Features

### Throw away your spreadsheet

Your experiments are all in one place, with filter and sort. Because the data's stored on S3, you can even see experiments that were run on other machines.

```shell-session
$ replicate ls --filter "val_loss<0.2"
EXPERIMENT   HOST         STATUS    BEST CHECKPOINT
e510303      10.52.2.23   stopped   49668cb (val_loss=0.1484)
9e97e07      10.52.7.11   running   41f0c60 (val_loss=0.1989)
```

### Analyze in a notebook

Don't like the CLI? No problem. You can retrieve, analyze, and plot your results from within a notebook. Think of it like a programmable Tensorboard.

<img src="web/public/images/notebook.png" width="600" />

### Compare experiments

It diffs everything, all the way down to versions of dependencies, just in case that latest Tensorflow version did something weird.

```shell-session
$ replicate diff 49668cb 41f0c60
Checkpoint:       49668cb     41f0c60
Experiment:       e510303     9e97e07

Params
learning_rate:    0.001       0.002

Python Packages
tensorflow:       2.3.0       2.3.1

Metrics
train_loss:       0.4626      0.8155
train_accuracy:   0.7909      0.7254
val_loss:         0.1484      0.1989
val_accuracy:     0.9607      0.9411
```

### Commit to Git, after the fact

There's no need to carefully commit everything to Git. Replicate lets you get back to any point you called `experiment.checkpoint()`, so you can commit to Git once you've found something that works.

```shell-session
$ replicate checkout f81069d
Copying code and weights to working directory...

# save the code to git
$ git commit -am "Use hinge loss"
```

### Load models in production

You can use Replicate to feed your models into production systems. Connect them back to how they were trained, who trained them, and what their metrics were.

```python
import replicate
model = torch.load(replicate.experiments.get("e45a203").best().open("model.pth"))
```

## Install

```
pip install -U replicate
```

## Get started

If you prefer **training scripts and the CLI**, [follow the our tutorial to learn how Replicate works](https://beta.replicate.ai/docs/tutorial).

If you prefer **working in notebooks**, <a href="https://colab.research.google.com/drive/1vjZReg--45P-NZ4j8TXAJFWuepamXc7K" target="_blank">follow our notebook tutorial on Colab</a>.

If you like to **learn concepts first**, [read our guide about how Replicate works](https://beta.replicate.ai/docs/learn/how-it-works).

## Contributing & development environment

[Take a look at our contributing instructions.](CONTRIBUTING.md)
