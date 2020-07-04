---
id: getting-started
title: Getting started
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

This guide will help you learn how Replicate works. We'll start with a model that we've written to show you the ropes, then you can have a go at implementing your model.

The goal of this tutorial is to make an image classifier that detects just ants and bees. It will be implemented as a pre-trained ResNet image classifier model, with the whole network frozen besides the final layer. The last layer is replaced by one with random weights and only this layer is trained. [Its source code is available here.](https://github.com/replicate/replicate-getting-started/blob/master/ant_bee_detector/model.py)

## Install Replicate

<Tabs
groupId="operating-systems"
defaultValue="mac"
values={[
{label: 'macOS', value: 'mac'},
{label: 'Linux', value: 'linux'},
]
}>
<TabItem value="mac">

Run the following commands in a terminal:

    curl -o /usr/local/bin/replicate https://storage.googleapis.com/replicate-public/cli/latest/darwin/amd64/replicate
    chmod +x /usr/local/bin/replicate

</TabItem>
<TabItem value="linux">
Run the following commands in a terminal:

    sudo curl -o /usr/local/bin/replicate https://storage.googleapis.com/replicate-public/cli/latest/linux/amd64/replicate
    sudo chmod +x /usr/local/bin/replicate

</TabItem>
</Tabs>

## Check out the model's repository

Check out the model we're working with:

    git clone https://github.com/replicate/replicate-getting-started.git
    cd replicate-getting-started/

You can also [browse the repository on GitHub](https://github.com/replicate/replicate-getting-started).

If you haven't got access to this repository, email [team@replicate.ai](mailto:team@replicate.ai).

## Start training

## What's next

So far, everything we've been doing has been local on a single machine. But in practice, you probably want to train on a separate machine with GPUs, or perhaps multiple machines at the same time. Or, perhaps you want to share your experiments with other people in your team.

Take a look at [the guide to working with remote machines](working-with-remote-machines.md) to learn more about this.
