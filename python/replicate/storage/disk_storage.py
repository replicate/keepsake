import os
from typing import AnyStr, Generator

from .storage_base import Storage, ListFileInfo
from ..exceptions import DoesNotExistError
from .._shared import lib
from ..shared_utils import (
    init_go_slice,
    py_str_to_go,
    go_bytes_to_py_bytes,
    go_str_to_py,
    init_go_string,
)


class DiskStorage(Storage):
    """
    Stores data on local filesystem
    """

    def __init__(self, root):
        self.root = os.path.join(os.getcwd(), root)

    def get(self, path: str) -> bytes:
        """
        Get data at path
        """
        ret_slice = init_go_slice()
        err_string = init_go_string()
        lib.DiskStorageGet(
            py_str_to_go(self.root)[0], py_str_to_go(path)[0], ret_slice, err_string[0]
        )
        print(err_string[1].n)
        return go_bytes_to_py_bytes(ret_slice)

    def put(self, path: str, data: AnyStr):
        """
        Save data to file at path
        """
        full_path = os.path.join(self.root, path)
        os.makedirs(os.path.dirname(full_path), exist_ok=True)

        mode = "w"
        if isinstance(data, bytes):
            mode = "wb"
        with open(full_path, mode) as fh:
            fh.write(data)

    def list(self, path: str) -> Generator[ListFileInfo, None, None]:
        """
        List files at path
        """
        # This is not recursive, but S3-style APIs make it very efficient to do recursive lists, so we probably want to add that
        full_path = os.path.join(self.root, path)
        for filename in os.listdir(full_path):
            if os.path.isfile(os.path.join(full_path, filename)):
                yield {"name": filename, "type": "file"}
            else:
                yield {"name": filename, "type": "directory"}

    def delete(self, path: str):
        """
        Delete single file at path
        """
        full_path = os.path.join(self.root, path)
        try:
            os.unlink(full_path)
        except FileNotFoundError:
            raise DoesNotExistError("No such path: '{}'".format(full_path))
