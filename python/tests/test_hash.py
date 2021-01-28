import random

from keepsake.hash import random_hash


def test_hash_length():
    assert len(random_hash(0)) == 0
    assert len(random_hash(1)) == 1
    assert len(random_hash(63)) == 63
    assert len(random_hash(64)) == 64
    assert len(random_hash(65)) == 65
    assert len(random_hash(1000)) == 1000


def test_hash_random_seed():
    random.seed(0)
    h1 = random_hash()
    random.seed(0)
    h2 = random_hash()
    assert h1 != h2


def test_immediate_hash():
    h1, h2 = random_hash(), random_hash()
    assert h1 != h2
