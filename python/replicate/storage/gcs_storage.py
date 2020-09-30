from typing import AnyStr, Generator

from .storage_base import Storage, ListFileInfo
from .. import shared
from ..exceptions import DoesNotExistError


class GCSStorage(Storage):
    def __init__(self, bucket: str, root: str):
        self.bucket_name = bucket
        self.root = root

    def root_url(self):
        """
        Returns the URL this storage is pointing at
        """
        ret = "gs://" + self.bucket_name
        if self.root:
            ret += "/" + self.root
        return ret

    def get(self, path: str) -> bytes:
        """
        Get data at path
        """
        try:
            result = shared.call(
                "GCSStorage.Get",
                Bucket=self.bucket_name,
                Root=self.root,
                Path=str(path),  # typecast for pathlib
            )
        except shared.SharedError as e:
            if e.type == "DoesNotExistError":
                raise DoesNotExistError(e.message)
            raise
        return result["Data"]

    def put_path(self, dest_path: str, source_path: str):
        """
        Save file or directory to path
        """
        shared.call(
            "GCSStorage.PutPath",
            Bucket=self.bucket_name,
            Root=self.root,
            Src=str(source_path),
            Dest=str(dest_path),
        )

    def put(self, path: str, data: AnyStr):
        """
        Save data to file at path
        """
        if isinstance(data, str):
            data_bytes = data.encode("utf-8")
        else:
            data_bytes = data
        shared.call(
            "GCSStorage.Put",
            Bucket=self.bucket_name,
            Root=self.root,
            Path=str(path),
            Data=data_bytes,
        )

    def delete(self, path: str):
        # TODO
        pass

    def list(self, path: str) -> Generator[ListFileInfo, None, None]:
        # TODO
        pass
