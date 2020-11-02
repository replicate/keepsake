import re
import replicate
import replicate.version


def test_version():
    assert re.match("^[0-9]+\.[0-9]+\.[0-9]+$", replicate.__version__)
    assert replicate.__version__ == replicate.version.version
