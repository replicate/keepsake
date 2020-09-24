# TODO(andreas): better error handling

import os
import re
from typing import AnyStr, Optional, Generator, Set, Any
import boto3  # type: ignore

from .storage_base import Storage, ListFileInfo
from .parallel_copy import Copier
from ..exceptions import DoesNotExistError


class S3Storage(Storage):
    """
    Stores data on Amazon S3
    """

    def __init__(self, bucket: str, root: str, concurrency=512):
        self.bucket_name = bucket
        self.root = root
        self.client = None
        self.concurrency = concurrency

    def get(self, path: str) -> bytes:
        """
        Get data at path
        """
        client = self._get_client()
        key = os.path.join(self.root, path)
        try:
            obj = client.get_object(Bucket=self.bucket_name, Key=key)
        except client.exceptions.NoSuchKey:
            raise DoesNotExistError()
        ret = obj["Body"].read()  # type: ignore
        return ret

    def put_path(self, path: str, source_path: str):
        """
        Save file or directory to s3 path
        """
        client = self._get_client()
        root_path = os.path.join(self.root, path)
        if os.path.isfile(source_path):
            with open(source_path, "rb") as fh:
                data = fh.read()
            client.put_object(Bucket=self.bucket_name, Key=root_path, Body=data)

        self.parallel_copy(S3ParallelCopier(self, path), source_path)

    def put(self, path: str, data: AnyStr):
        """
        Save data to file at path
        """
        key = os.path.join(self.root, path)
        if isinstance(data, str):
            data_bytes = data.encode()
        else:
            data_bytes = data

        client = self._get_client()
        try:
            client.put_object(Bucket=self.bucket_name, Key=key, Body=data_bytes)
        except client.exceptions.NoSuchBucket:
            self.create_bucket()
            client.put_object(Bucket=self.bucket_name, Key=key, Body=data_bytes)

    def list(self, path: str) -> Generator[ListFileInfo, None, None]:
        """
        List files at path
        """
        # TODO(andreas): inefficiently fetches all paths in bucket

        client = self._get_client()
        paginator = client.get_paginator("list_objects")
        prefix = os.path.join(self.root, path)

        if not prefix.endswith("/"):
            prefix += "/"

        if prefix.startswith("/"):
            prefix = prefix[1:]

        rel_path_regex = re.compile("^" + re.escape(prefix))

        seen_dirs: Set[str] = set()
        for result in paginator.paginate(Bucket=self.bucket_name, Prefix=prefix,):
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
        client = self._get_client()
        key = os.path.join(self.root, path)
        try:
            client.head_object(Bucket=self.bucket_name, Key=key)
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

        client = self._get_client()
        key = os.path.join(self.root, path)
        client.delete_object(Bucket=self.bucket_name, Key=key)

    def _get_client(self):
        if self.client is not None:
            return self.client

        self.client = boto3.client("s3")  # type: ignore
        return self.client

    def create_bucket(self):
        self._get_client().create_bucket(Bucket=self.bucket_name)
        bucket = boto3.resource("s3").Bucket(self.bucket_name)
        bucket.wait_until_exists()


class S3ParallelCopier(Copier):

    def __init__(self, storage: S3Storage, path):
        super().__init__()

        self.storage = storage
        self.path = path
        self.client = None

    def begin(self):
        self.client = boto3.client("s3")

    def copy(self, rel_path: str, abs_path: str):
        assert self.client is not None, "begin() must be called before copy()"

        root_path = os.path.join(self.storage.root, self.path)
        remote_path = os.path.join(root_path, rel_path)

        with open(abs_path, "rb") as f:
            data = f.read()

        print("****************", self.storage.bucket_name, self.path, remote_path, len(data), rel_path, abs_path)

        self.client.put_object(Bucket=self.storage.bucket_name, Key=remote_path, Body=data)
