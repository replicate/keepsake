import os
from urllib.parse import urlparse

from .storage_base import Storage
from .disk_storage import DiskStorage

from ..exceptions import UnknownStorageBackend


def storage_for_url(url: str) -> Storage:
    parsed_url = urlparse(url)

    if parsed_url.scheme == "" or parsed_url.scheme == "file":
        # don't use os.path.join() here because path starts with "/" and join will treat that as root URL
        return DiskStorage(root=parsed_url.netloc + parsed_url.path)
    elif parsed_url.scheme == "s3":
        # lazy import to speed up import replicate
        from .s3_storage import S3Storage

        return S3Storage(bucket=parsed_url.netloc, root=parsed_url.path.lstrip("/"))
    elif parsed_url.scheme == "gs":
        from .gcs_storage import GCSStorage

        return GCSStorage(bucket=parsed_url.netloc, root=parsed_url.path.lstrip("/"))
    else:
        raise UnknownStorageBackend(parsed_url.scheme)
