import re
import keepsake
import keepsake.version


def test_version():
    assert re.match("^[0-9]+\.[0-9]+\.[0-9]+$", keepsake.__version__)
    assert keepsake.__version__ == keepsake.version.version
