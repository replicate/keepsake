import os
import asyncio
from typing import AnyStr, Optional, Generator, Set
import aiohttp
from gcloud.aio.storage import Storage as AioStorage  # type: ignore
from google.cloud import storage  # type: ignore
from google.api_core import exceptions  # type: ignore

from .storage_base import Storage, ListFileInfo
from ..exceptions import DoesNotExistError


class GCSStorage(Storage):
    def __init__(self, bucket: str, concurrency=512):
        self.bucket_name = bucket
        self.client: Optional[storage.Client] = None
        self.concurrency = concurrency

    def get(self, path: str) -> bytes:
        """
        Get data at path
        """
        bucket = self.bucket()
        try:
            blob = bucket.blob(path)
            return blob.download_as_string()
        except exceptions.NotFound:
            raise DoesNotExistError()

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
        async with aiohttp.ClientSession() as session:
            storage = AioStorage(session=session)
            for relative_path, data in self.walk_directory_data(dir_to_store):
                remote_path = os.path.join(path, relative_path)

                put_task = loop.create_task(
                    storage.upload(self.bucket_name, remote_path, data)
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

        bucket = self.bucket()
        blob = bucket.blob(path)
        blob.upload_from_string(data)

    def get_client(self) -> storage.Client:
        if self.client is None:
            self.client = storage.Client()
        return self.client

    def bucket(self) -> storage.Bucket:
        """
        Return a GCS storage.Bucket object. If the bucket doesn't
        exist, it's automatically created.
        """
        client = self.get_client()
        try:
            return client.get_bucket(self.bucket_name)
        except exceptions.NotFound:
            client.create_bucket(self.bucket_name)
            return client.get_bucket(self.bucket_name)

    def delete(self, path: str):
        # TODO
        pass

    def list(self, path: str) -> Generator[ListFileInfo, None, None]:
        # TODO
        pass
