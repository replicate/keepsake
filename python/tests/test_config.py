import pytest  # type: ignore
import os

from replicate.config import (
    load_config,
    validate_and_set_defaults,
    ConfigValidationError,
)


def test_load_config_blank(tmp_path):
    config_file = tmp_path / "replicate.yaml"
    config_file.write_text("")

    with pytest.raises(ConfigValidationError):
        load_config(tmp_path)


def test_validate():
    validate_and_set_defaults({"repository": "s3://foobar"}, "/foo")
    with pytest.raises(ConfigValidationError):
        validate_and_set_defaults({"invalid": "key"}, "/foo")
    with pytest.raises(ConfigValidationError):
        validate_and_set_defaults({"repository": 1234}, "/foo")
    with pytest.raises(ConfigValidationError):
        validate_and_set_defaults(
            {"repository": "s3://foobar", "something": "else"}, "/foo"
        )

    assert validate_and_set_defaults({"repository": "s3://foobar"}, "/foo") == {
        "repository": "s3://foobar",
    }


def test_storage_backwards_compatible():
    assert validate_and_set_defaults({"storage": "s3://foobar"}, "/foo") == {
        "repository": "s3://foobar",
    }
    with pytest.raises(ConfigValidationError):
        validate_and_set_defaults(
            {"storage": "s3://foobar", "repository": "s3://foobar"}, "/foo"
        )
