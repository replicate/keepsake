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

    assert load_config(tmp_path) == {
        "python": "3.7",
        "storage": os.path.join(tmp_path, ".replicate/storage/"),
    }


def test_validate():
    validate_and_set_defaults({"storage": "s3://foobar"}, "/foo")
    with pytest.raises(ConfigValidationError):
        validate_and_set_defaults({"invalid": "key"}, "/foo")
        validate_and_set_defaults({"storage": 1234}, "/foo")
        validate_and_set_defaults(
            {"storage": "s3://foobar", "something": "else"}, "/foo"
        )

    assert validate_and_set_defaults({}, "/foo") == {
        "python": "3.7",
        "storage": "/foo/.replicate/storage/",
    }
