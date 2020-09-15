import tempfile
import json
import sys
import base64
import binascii
import os
from typing import AnyStr, Optional, Generator, Set, Tuple
from google.cloud import storage  # type: ignore
from google.api_core import exceptions  # type: ignore
from google.auth.credentials import Credentials  # type: ignore
from google.oauth2 import service_account  # type: ignore

from .storage_base import Storage, ListFileInfo
from ..exceptions import DoesNotExistError


class GCSStorage(Storage):
    def __init__(self, bucket: str, root: str, concurrency=512):
        self.bucket_name = bucket
        self.root = root
        self.client: Optional[storage.Client] = None
        self.concurrency = concurrency
        self.temp_service_account_path: Optional[str] = None

    def get(self, path: str) -> bytes:
        """
        Get data at path
        """
        bucket = self.bucket()
        try:
            blob = bucket.blob(os.path.join(self.root, path))
            return blob.download_as_string()
        except exceptions.NotFound:
            raise DoesNotExistError()

    def put_directory(self, path: str, dir_to_store: str):
        """
        Save directory to path
        """
        bucket = self.bucket()
        for relative_path, absolute_path in self.walk_directory_paths(dir_to_store):
            remote_path = os.path.join(self.root, path, relative_path)

            if os.environ.get("REPLICATE_DEBUG"):
                sys.stderr.write(
                    "Uploading to gs://{}/{}\n".format(self.bucket_name, remote_path)
                )
                sys.stderr.flush()

            blob = bucket.blob(remote_path)
            blob.upload_from_filename(absolute_path)

    def put(self, path: str, data: AnyStr):
        """
        Save data to file at path
        """
        bucket = self.bucket()
        remote_path = os.path.join(self.root, path)
        if os.environ.get("REPLICATE_DEBUG"):
            sys.stderr.write(
                "Uploading to gs://{}/{}\n".format(self.bucket_name, remote_path)
            )
            sys.stderr.flush()
        blob = bucket.blob(remote_path)
        blob.upload_from_string(data)

    def _get_client(self) -> storage.Client:
        if self.client is None:
            project, credentials = self._get_project_and_client_credentials()
            if project is None:
                return storage.Client()
            else:
                self.client = storage.Client(project=project, credentials=credentials)
        return self.client

    def bucket(self) -> storage.Bucket:
        """
        Return a GCS storage.Bucket object. If the bucket doesn't
        exist, it's automatically created.
        """
        client = self._get_client()
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

    def _get_service_account_key(self) -> Optional[bytes]:
        key_b64 = os.environ.get("REPLICATE_GCP_SERVICE_ACCOUNT_KEY")
        if key_b64 is None:
            return None
        try:
            key_json = base64.standard_b64decode(key_b64)
        except binascii.Error as e:
            # TODO(andreas): fail hard here?
            # TODO(andreas): actual logging facility
            sys.stderr.write(
                "Failed to decode service account key from base-64: {}".format(e)
            )
            return None

        return key_json

    def _get_project_and_client_credentials(
        self,
    ) -> Tuple[Optional[str], Optional[Credentials]]:
        key_json = self._get_service_account_key()
        if key_json is None:
            return None, None

        try:
            info = json.loads(key_json)
        except json.JSONDecodeError as e:
            # TODO(andreas): fail hard here?
            sys.stderr.write("Failed to decode service account key JSON: {}".format(e))
            return None, None

        project = os.environ.get("REPLICATE_GCP_PROJECT")
        return project, service_account.Credentials.from_service_account_info(info)
