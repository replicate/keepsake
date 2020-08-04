import time
import torch
import replicate


print(torch.cuda.device_count())

experiment = replicate.init(learning_rate=0.002)

print("Starting experiment {}...".format(experiment.id))

for epoch in range(10):
    print("Committing epoch {}...".format(epoch))
    experiment.commit(step=epoch, accuracy="0.987")
    time.sleep(1)
