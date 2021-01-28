import hashlib
import time
import random


def random_hash(length=64) -> str:
    # time-dependent hash to begin with, then pad with random characters
    hasher = hashlib.sha256()
    hasher.update(str(time.time()).encode("utf-8"))
    hasher.update(str(random.random()).encode("utf-8"))
    digest = hasher.hexdigest()
    if len(digest) >= length:
        return digest[:length]
    padding = "".join(
        random.choice("0123456789abcdef") for i in range(length - len(digest))
    )
    return digest + padding
