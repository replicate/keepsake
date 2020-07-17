import sys
import os
from abc import ABCMeta, abstractmethod
from typing import AnyStr, Generator

if sys.version_info >= (3, 8):
    from typing import TypedDict
else:
    from typing_extensions import TypedDict

ListFileInfo = TypedDict("ListFileInfo", {"name": str, "type": str})


class Storage:
    __metaclass__ = ABCMeta

    @abstractmethod
    def get(self, path: str) -> bytes:
        """
        Get data at path
        """
        raise NotImplementedError()

    @abstractmethod
    def put(self, path: str, data: AnyStr):
        """
        Save data to file at path
        """
        raise NotImplementedError()

    def put_directory(self, path: str, dir_to_store: str):
        """
        Save directory to path

        Parallels storage.PutDirectory in Go.
        """
        ignore = [".replicate", ".git", "venv"]

        for current_directory, dirs, files in os.walk(dir_to_store, topdown=True):
            dirs[:] = [d for d in dirs if d not in ignore]

            for filename in files:
                with open(os.path.join(current_directory, filename), "rb") as fh:
                    data = fh.read()
                # Strip local path
                relative_path = os.path.join(
                    os.path.relpath(current_directory, dir_to_store), filename
                )
                # Then, make it relative to path we want to store it in storage
                self.put(os.path.join(path, relative_path), data)

    @abstractmethod
    def list(self, path: str) -> Generator[ListFileInfo, None, None]:
        """
        List files at path
        """
        raise NotImplementedError()

    @abstractmethod
    def delete(self, path: str):
        """
        Delete single file at path
        """
        raise NotImplementedError()
