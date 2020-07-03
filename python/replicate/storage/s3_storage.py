# TODO(andreas): better error handling

import re
from typing import AnyStr, Optional, Generator, Set
import boto3
import mypy_boto3_s3 as s3

from .storage_base import Storage, ListFileInfo
from ..exceptions import DoesNotExistError


class S3Storage(Storage):
    """
    Stores data on Amazon S3
    """

    def __init__(self, bucket: str):
        self.bucket = bucket
        self.client: Optional[s3.Client] = None

    def get(self, path: str) -> bytes:
        """
        Get data at path
        """
        client = self.get_client()
        try:
            obj = client.get_object(Bucket=self.bucket, Key=path)
        except client.exceptions.NoSuchKey:
            raise DoesNotExistError()
        ret = obj["Body"].read()
        return ret

    def put(self, path: str, data: AnyStr):
        """
        Save data to file at path
        """
        if isinstance(data, str):
            data_bytes = data.encode()
        else:
            data_bytes = data

        client = self.get_client()
        client.put_object(Bucket=self.bucket, Key=path, Body=data_bytes)

    def list(self, path: str) -> Generator[ListFileInfo, None, None]:
        """
        List files at path
        """
        # TODO(andreas): inefficiently fetches all paths in bucket

        client = self.get_client()
        paginator = client.get_paginator("list_objects")

        if not path.endswith("/"):
            path += "/"

        if path.startswith("/"):
            path = path[1:]

        rel_path_regex = re.compile("^" + re.escape(path))

        seen_dirs: Set[str] = set()
        for result in paginator.paginate(Bucket=self.bucket, Prefix=path,):
            for content in result.get("Contents", []):
                object_path = content["Key"]
                rel_path = rel_path_regex.sub("", object_path)
                path_parts = rel_path.split("/")
                if len(path_parts) > 1:
                    dir_path = path_parts[0]
                    if dir_path not in seen_dirs:
                        seen_dirs.add(dir_path)
                        yield {"name": dir_path, "type": "directory"}
                else:
                    yield {"name": rel_path, "type": "file"}

    def exists(self, path: str) -> bool:
        client = self.get_client()
        try:
            client.head_object(Bucket=self.bucket, Key=path)
        except client.exceptions.ClientError as e:
            code = e.response.get("Error", {}).get("Code")
            if code == "404":
                return False
            raise
        return True

    def delete(self, path: str):
        """
        Delete single file at path
        """
        if not self.exists(path):
            raise DoesNotExistError()

        client = self.get_client()
        client.delete_object(Bucket=self.bucket, Key=path)

    def get_client(self) -> s3.Client:
        if self.client is not None:
            return self.client

        self.client = boto3.client("s3")
        return self.client
