import datetime
from keepsake.packages import get_imported_packages


def test_get_imported_packages():
    assert "keepsake" in get_imported_packages()
