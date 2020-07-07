import replicate


experiment = replicate.init(params={"learning_rate": 0.002})

print("Starting experiment {}...".format(experiment.id))

for epoch in range(10):
    print("Committing epoch {}...".format(epoch))
    experiment.commit(metrics={"epoch": epoch})
