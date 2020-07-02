import replicate


experiment = replicate.init(params={"learning_rate": 0.002})

print("Starting experiment " + experiment.id)

for epoch in range(10):
    experiment.commit(metrics={"epoch": epoch})

