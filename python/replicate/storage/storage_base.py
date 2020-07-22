import sys
import os
from abc import ABCMeta, abstractmethod
from typing import AnyStr, Generator, Tuple

if sys.version_info >= (3, 8):
    from typing import TypedDict
else:
    from typing_extensions import TypedDict

ListFileInfo = TypedDict("ListFileInfo", {"name": str, "type": str})


class Storage:
    __metaclass__ = ABCMeta

    put_directory_ignore = [".replicate", ".git", "venv", ".mypy_cache"]

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
        for relative_path, data in self.walk_directory_data(dir_to_store):
            # Make it relative to path we want to store it in storage
            self.put(os.path.join(path, relative_path), data)

    def walk_directory_data(
        self, directory: str
    ) -> Generator[Tuple[str, bytes], None, None]:
        """
        Yields (relative_path, data) of all files, recursively, in
        directory.
        """
        for current_directory, dirs, files in os.walk(directory, topdown=True):
            dirs[:] = [d for d in dirs if d not in self.put_directory_ignore]

            for filename in files:
                with open(os.path.join(current_directory, filename), "rb") as fh:
                    data = fh.read()
                # Strip local path
                relative_dir = os.path.relpath(current_directory, directory)
                # relative_dir will be "." if current_directory ==
                # dir_to_store, this period will be added to the
                # bucket path
                if relative_dir == ".":
                    relative_dir = ""
                relative_path = os.path.join(relative_dir, filename)

                yield relative_path, data

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
