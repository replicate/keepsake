import datetime
from replicate.packages import get_imported_packages


def test_get_imported_packages():
    assert "replicate" in get_imported_packages()
