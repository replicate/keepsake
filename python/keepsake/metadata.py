import datetime


def rfc3339_datetime(dt: datetime.datetime) -> str:
    """
    Return a datetime in RFC3339 format.

    datetime.utcnow().isoformat() is so close -- it's just missing the UTC timezone suffix!
    Go requires this to be able to parse it using the default time.Time parser.

    datetime.now(timezone.utc) is almost it, but uses +00:00 instead of the Z shorthand.

    https://bugs.python.org/issue35829 for the inverse problem.
    """
    assert (
        dt.tzinfo is None
    ), "rfc3339_datetime() only works with naive datetime objects"
    return dt.isoformat() + "Z"


def parse_rfc3339(s: str) -> datetime.datetime:
    """
    Parse a string in the RFC3339 format we use.
    """
    return datetime.datetime.strptime(s, "%Y-%m-%dT%H:%M:%S.%fZ")
