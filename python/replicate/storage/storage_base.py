import sys
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

    def put_path(self, dest_path: str, source_path: str):
        """
        Save file or directory to path on storage
        """
        raise NotImplementedError()

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
