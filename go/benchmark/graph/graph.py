#!/usr/bin/env python

import fileinput
import re

import numpy as np
import matplotlib.pyplot as plt
from sklearn.linear_model import LinearRegression

graphs = {}

for line in fileinput.input():
    # BenchmarkKeepsakeDisk/list_first_run_with_10_experiments-8         	      10	  20338507 ns/op
    match = re.search(
        r"BenchmarkKeepsake(.+)\/(.+)_with_(\d+).+\s+\d+\s+(\d+) ns\/op", line
    )
    if match:
        graph = match.group(1) + " " + match.group(2)
        num_experiments = int(match.group(3))
        time = int(match.group(4)) * 1e-9  # nanoseconds to seconds

        if graph not in graphs:
            graphs[graph] = {"x": [], "y": []}

        graphs[graph]["x"].append(num_experiments)
        graphs[graph]["y"].append(time)


for graph_name, data in graphs.items():
    x = np.array(data["x"]).reshape((-1, 1))
    y = np.array(data["y"])
    model = LinearRegression()
    model.fit(x, y)
    x_new = np.linspace(0, 100, 100)
    y_new = model.predict(x_new[:, np.newaxis])

    plt.figure(figsize=(10, 7))

    ax = plt.axes()
    ax.scatter(x, y)
    ax.plot(x_new, y_new)

    ax.set_xlabel("number of experiments")
    ax.set_ylabel("time (s)")

    ax.axis("tight")
    ax.set_ylim(ymin=0)

    plt.title(graph_name)
    plt.savefig(graph_name + ".png")
