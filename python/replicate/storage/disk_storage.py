import os
from typing import AnyStr, List

from .storage_base import Storage
from .. import shared
from ..exceptions import DoesNotExistError


class DiskStorage(Storage):
    """
    Stores data on local filesystem

    Unlike the remote storages, some of these methods are implemented natively
    because they're trivial. The complex and slow ones (e.g. put_path) we call Go.
    """

    def __init__(self, root):
        self.root = root

    def root_url(self):
        """
        Returns the path this storage is pointing at
        """
        return self.root

    def get(self, path: str) -> bytes:
        """
        Get data at path
        """
        full_path = os.path.join(self.root, path)
        try:
            with open(full_path, "rb") as fh:
                return fh.read()
        except FileNotFoundError:
            raise DoesNotExistError("No such path: '{}'".format(full_path))

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

    def put_path(self, source_path: str, dest_path: str):
        """
        Save file or directory to path
        """
        shared.call(
            "DiskStorage.PutPath",
            Root=self.root,
            Src=str(source_path),
            Dest=str(dest_path),
        )

    def put_path_tar(self, local_path: str, tar_path: str, include_path: str):
        """
        Save file or directory to tarball
        """
        shared.call(
            "DiskStorage.PutPathTar",
            Root=self.root,
            LocalPath=str(local_path),
            TarPath=str(tar_path),
            IncludePath=str(include_path),
        )

    def list(self, path: str) -> List[str]:
        """
        Returns a list of files at path, but not any subdirectories.

        Returned paths are prefixed with the given path, that can be passed straight to Get().
        Directories are not listed.
        If path does not exist, an empty list will be returned.
        """
        full_path = os.path.join(self.root, path)
        result: List[str] = []
        for filename in os.listdir(full_path):
            if os.path.isfile(os.path.join(full_path, filename)):
                result.append(os.path.join(path, filename))
        return result

    def delete(self, path: str):
        """
        Recursively delete path
        """
        # Even though it's a simple operation we use the shared
        # library to ensure consistent semantics.
        shared.call(
            "DiskStorage.Delete",
            Root=self.root,
            Path=str(path),  # typecast for pathlib
        )
