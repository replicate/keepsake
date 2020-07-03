import pytest  # type: ignore

from replicate.config import validate_and_set_defaults, ConfigValidationError


def test_validate():
    validate_and_set_defaults({"storage": "s3://foobar"})
    with pytest.raises(ConfigValidationError):
        validate_and_set_defaults({"invalid": "key"})
        validate_and_set_defaults({"storage": 1234})
        validate_and_set_defaults({"storage": "s3://foobar", "something": "else"})

    assert validate_and_set_defaults({}) == {"storage": ".replicate/storage/"}
