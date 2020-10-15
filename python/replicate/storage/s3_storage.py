from typing import AnyStr, List

from .storage_base import Storage
from .. import shared
from ..exceptions import DoesNotExistError


class S3Storage(Storage):
    """
    Stores data on Amazon S3
    """

    def __init__(self, bucket: str, root: str):
        self.bucket_name = bucket
        self.root = root

    def root_url(self):
        """
        Returns the URL this storage is pointing at
        """
        ret = "s3://" + self.bucket_name
        if self.root:
            ret += "/" + self.root
        return ret

    def get(self, path: str) -> bytes:
        """
        Get data at path
        """
        try:
            result = shared.call(
                "S3Storage.Get",
                Bucket=self.bucket_name,
                Root=self.root,
                Path=str(path),  # typecast for pathlib
            )
        except shared.SharedError as e:
            if e.type == "DoesNotExistError":
                raise DoesNotExistError(e.message)
            raise
        return result["Data"]

    def put_path(self, source_path: str, dest_path: str):
        """
        Save directory to path
        """
        shared.call(
            "S3Storage.PutPath",
            Bucket=self.bucket_name,
            Root=self.root,
            Src=str(source_path),
            Dest=str(dest_path),
        )

    def put_path_tar(self, local_path: str, tar_path: str, include_path: str):
        """
        Save file or directory to tarball
        """
        shared.call(
            "S3Storage.PutPathTar",
            Bucket=self.bucket_name,
            Root=self.root,
            LocalPath=str(local_path),
            TarPath=str(tar_path),
            IncludePath=str(include_path),
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
            "S3Storage.Put",
            Bucket=self.bucket_name,
            Root=self.root,
            Path=str(path),
            Data=data_bytes,
        )

    def list(self, path: str) -> List[str]:
        """
        Returns a list of files at path, but not any subdirectories.
        """
        result = shared.call(
            "S3Storage.List",
            Bucket=self.bucket_name,
            Root=self.root,
            Path=str(path),  # typecast for pathlib
        )
        return result["Paths"]

    def exists(self, path: str) -> bool:
        pass

    def delete(self, path: str):
        pass
