import unittest

from replicate.config import validate_and_set_defaults, ConfigValidationError


class TestConfig(unittest.TestCase):
    def test_validate(self):
        validate_and_set_defaults({"storage": "s3://foobar"})
        with self.assertRaises(ConfigValidationError):
            validate_and_set_defaults({"invalid": "key"})
            validate_and_set_defaults({"storage": 1234})
            validate_and_set_defaults({"storage": "s3://foobar", "something": "else"})
        
        self.assertEqual(validate_and_set_defaults({}), {"storage": ".replicate/storage/"})
