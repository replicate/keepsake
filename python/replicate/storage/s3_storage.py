# TODO(andreas): better error handling

import os
import asyncio
import re
from typing import AnyStr, Optional, Generator, Set
import aiobotocore
import boto3
import mypy_boto3_s3 as s3

from .storage_base import Storage, ListFileInfo
from ..exceptions import DoesNotExistError


class S3Storage(Storage):
    """
    Stores data on Amazon S3
    """

    def __init__(self, bucket: str, concurrency=512):
        self.bucket = bucket
        self.client: Optional[s3.Client] = None
        self.concurrency = concurrency

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

    def put_directory(self, path: str, dir_to_store: str):
        """
        Save directory to path
        """
        loop = asyncio.get_event_loop()
        # TODO(andreas): handle exceptions
        loop.run_until_complete(self.put_directory_async(loop, path, dir_to_store))

    async def put_directory_async(self, loop: asyncio.BaseEventLoop, path: str, dir_to_store: str):
        put_tasks = set()
        session = aiobotocore.get_session()
        async with session.create_client("s3") as client:
            for current_directory, dirs, files in os.walk(dir_to_store, topdown=True):
                dirs[:] = [d for d in dirs if d not in self.put_directory_ignore]

                for filename in files:
                    local_path = os.path.join(current_directory, filename)
                    # Strip local path
                    relative_path = os.path.join(
                        os.path.relpath(current_directory, dir_to_store), filename
                    )
                    # Then, make it relative to path we want to store it in storage
                    remote_path = os.path.join(path, relative_path)

                    with open(local_path, "rb") as f:
                        body = f.read()
                    put_task = loop.create_task(client.put_object(
                        Body=body, Bucket=self.bucket, Key=remote_path
                    ))

                    # Emulate a worker pool by waiting for a single
                    # task to finish when the number of tasks ==
                    # self.concurrency
                    put_tasks.add(put_task)
                    if len(put_tasks) >= self.concurrency:
                        _, put_tasks = await asyncio.wait(
                            put_tasks, return_when=asyncio.FIRST_COMPLETED
                        )

            await asyncio.wait(put_tasks)

    def put(self, path: str, data: AnyStr):
        """
        Save data to file at path
        """
        if isinstance(data, str):
            data_bytes = data.encode()
        else:
            data_bytes = data

        client = self.get_client()
        try:
            client.put_object(Bucket=self.bucket, Key=path, Body=data_bytes)
        except client.exceptions.NoSuchBucket:
            self.create_bucket()
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

    def create_bucket(self):
        self.get_client().create_bucket(Bucket=self.bucket)
        bucket = boto3.resource("s3").Bucket(self.bucket)
        bucket.wait_until_exists()
