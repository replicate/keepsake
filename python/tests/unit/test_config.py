import pytest  # type: ignore

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
        "storage": ".replicate/storage/",
    }


def test_validate():
    validate_and_set_defaults({"storage": "s3://foobar"})
    with pytest.raises(ConfigValidationError):
        validate_and_set_defaults({"invalid": "key"})
        validate_and_set_defaults({"storage": 1234})
        validate_and_set_defaults({"storage": "s3://foobar", "something": "else"})

    assert validate_and_set_defaults({}) == {
        "python": "3.7",
        "storage": ".replicate/storage/",
    }
