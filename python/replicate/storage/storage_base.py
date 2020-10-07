import sys
from abc import ABCMeta, abstractmethod
from typing import AnyStr, List


class Storage:
    __metaclass__ = ABCMeta

    @abstractmethod
    def root_url(self) -> str:
        """
        Returns the path or URL this storage is pointing at
        """
        raise NotImplementedError()

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
    def list(self, path: str) -> List[str]:
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
