import tempfile
import unittest

from replicate.storage import DiskStorage, DoesNotExistError


class TestDiskStorage(unittest.TestCase):
    def test_put_get(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            storage = DiskStorage(root=tmpdir)
            storage.put("some/file", "nice")
            self.assertEqual(storage.get("some/file"), "nice")
            with self.assertRaises(DoesNotExistError):
                storage.get("not/here")

    def test_list(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            storage = DiskStorage(root=tmpdir)
            storage.put("foo", "nice")
            storage.put("some/bar", "nice")
            self.assertEqual(
                list(storage.list("")),
                [
                    {"name": "foo", "type": "file"},
                    {"name": "some", "type": "directory"},
                ],
            )
            self.assertEqual(
                list(storage.list("some")), [{"name": "bar", "type": "file"}]
            )

    def test_delete(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            storage = DiskStorage(root=tmpdir)
            storage.put("some/file", "nice")
            self.assertEqual(storage.get("some/file"), "nice")
            storage.delete("some/file")
            with self.assertRaises(DoesNotExistError):
                storage.get("some/file")
