import sys
import os
from abc import ABCMeta, abstractmethod
from typing import AnyStr, Generator, Tuple
from gitignore_parser import parse_gitignore  # type: ignore

if sys.version_info >= (3, 8):
    from typing import TypedDict
else:
    from typing_extensions import TypedDict

ListFileInfo = TypedDict("ListFileInfo", {"name": str, "type": str})


class Storage:
    __metaclass__ = ABCMeta

    put_directory_ignore = [
        ".replicate",
        ".git",
        "venv",
        ".mypy_cache",
    ]

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

    def put_path(self, dest_path: str, source_path: str):
        """
        Save file or directory to path on storage

        Parallels storage.PutPath in Go.
        """
        if os.path.isfile(source_path):
            with open(source_path, "rb") as fh:
                data = fh.read()
            self.put(dest_path, data)
            return

        for relative_path, data in self.walk_directory_data(source_path):
            # Make it relative to path we want to store it in storage
            self.put(os.path.join(dest_path, relative_path), data)

    def walk_directory_paths(
        self, directory: str
    ) -> Generator[Tuple[str, str], None, None]:
        """
        Yields (relative_path, absolute_path) of all files,
        recursively, in directory.
        """
        ignorefile_path = os.path.join(directory, ".replicateignore")
        if os.path.exists(ignorefile_path):
            ignore_matches = parse_gitignore(ignorefile_path)
        else:
            ignore_matches = None

        for current_directory, dirs, files in os.walk(directory, topdown=True):
            dirs[:] = [d for d in dirs if d not in self.put_directory_ignore]

            for filename in files:
                absolute_path = os.path.join(current_directory, filename)
                # Strip local path
                relative_dir = os.path.relpath(current_directory, directory)
                # relative_dir will be "." if current_directory ==
                # dir_to_store, this period will be added to the
                # bucket path
                if relative_dir == ".":
                    relative_dir = ""
                relative_path = os.path.join(relative_dir, filename)

                if ignore_matches is not None and ignore_matches(absolute_path):
                    continue

                yield relative_path, absolute_path

    def walk_directory_data(
        self, directory: str
    ) -> Generator[Tuple[str, bytes], None, None]:
        """
        Yields (relative_path, data) of all files, recursively, in
        directory.
        """
        for relative_path, absolute_path in self.walk_directory_paths(directory):
            with open(absolute_path, "rb") as fh:
                data = fh.read()
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
