# TODO(andreas): better error handling

import os
import asyncio
import re
from typing import AnyStr, Optional, Generator, Set, Any
import aiobotocore  # type: ignore
import boto3
import mypy_boto3_s3 as s3

from .storage_base import Storage, ListFileInfo
from ..exceptions import DoesNotExistError


class S3Storage(Storage):
    """
    Stores data on Amazon S3
    """

    def __init__(self, bucket: str, concurrency=512):
        self.bucket_name = bucket
        self.client: Optional[s3.Client] = None
        self.concurrency = concurrency

    def get(self, path: str) -> bytes:
        """
        Get data at path
        """
        client = self.get_client()
        try:
            obj = client.get_object(Bucket=self.bucket_name, Key=path)
        except client.exceptions.NoSuchKey:
            raise DoesNotExistError()
        ret = obj["Body"].read()  # type: ignore
        return ret

    def put_directory(self, path: str, dir_to_store: str):
        """
        Save directory to path
        """
        loop = asyncio.get_event_loop()
        # TODO(andreas): handle exceptions
        loop.run_until_complete(self.put_directory_async(loop, path, dir_to_store))

    async def put_directory_async(
        self, loop: asyncio.AbstractEventLoop, path: str, dir_to_store: str
    ):
        put_tasks = set()
        session = aiobotocore.get_session()
        async with session.create_client("s3") as client:
            for relative_path, body in self.walk_directory_data(dir_to_store):
                remote_path = os.path.join(path, relative_path)

                put_task = loop.create_task(
                    client.put_object(
                        Body=body, Bucket=self.bucket_name, Key=remote_path
                    )
                )

                # Emulate a worker pool by waiting for a single task
                # to finish when the number of tasks == self.concurrency
                put_tasks.add(put_task)
                if len(put_tasks) >= self.concurrency:
                    _, new_tasks = await asyncio.wait(
                        put_tasks, return_when=asyncio.FIRST_COMPLETED
                    )
                    for task in new_tasks:
                        put_tasks.add(loop.create_task(task))

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
            client.put_object(Bucket=self.bucket_name, Key=path, Body=data_bytes)
        except client.exceptions.NoSuchBucket:
            self.create_bucket()
            client.put_object(Bucket=self.bucket_name, Key=path, Body=data_bytes)

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
        for result in paginator.paginate(Bucket=self.bucket_name, Prefix=path,):
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
            client.head_object(Bucket=self.bucket_name, Key=path)
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
        client.delete_object(Bucket=self.bucket_name, Key=path)

    def get_client(self) -> s3.Client:
        if self.client is not None:
            return self.client

        self.client = boto3.client("s3")  # type: ignore
        return self.client

    def create_bucket(self):
        self.get_client().create_bucket(Bucket=self.bucket_name)
        bucket = boto3.resource("s3").Bucket(self.bucket_name)
        bucket.wait_until_exists()
