# Example model

A simple neural net to classify Iris flowers in the [Iris data set](https://archive.ics.uci.edu/ml/datasets/iris). It is the example used in [the Replicate tutorial](https://beta.replicate.ai/docs/tutorial/).

This is a somewhat contrived model to demonstrate how Replicate works. It's main requirement is that it trains fast, so you can try out Replicate as quickly as possible!

## Install the dependencies

Install the dependencies inside a Virtualenv:

    virtualenv venv
    source venv/bin/activate
    pip install -r requirements.txt

## Train

    python train.py
