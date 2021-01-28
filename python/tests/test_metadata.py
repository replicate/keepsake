import datetime
from keepsake.metadata import parse_rfc3339


def test_parse_rfc3339():
    dt = parse_rfc3339("2020-10-07T22:44:06.243914Z")
    assert dt == datetime.datetime(2020, 10, 7, 22, 44, 6, 243914)
    assert dt.tzinfo is None
