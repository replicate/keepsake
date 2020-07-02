import random


def random_hash(length=64):
    return "".join(random.choice("0123456789abcdef") for i in range(length))
