import re

from .storage_base import Storage
from .disk_storage import DiskStorage
from .s3_storage import S3Storage


class DoesNotExistError(Exception):
    pass


class UnknownStorageBackend(Exception):
    def __init__(self, scheme):
        super(self, UnknownStorageBackend).__init__(
            "Unknown storage backend: {}".format(scheme)
        )


def storage_for_url(url: str) -> Storage:
    url_re = re.compile("^([^:]+)://(.+)$")
    match = url_re.match(url)
    if match is None:
        return DiskStorage(root=url)

    scheme, path = match.groups()

    if scheme == "s3://":
        return S3Storage(bucket=path)
    else:
        raise UnknownStorageBackend(scheme)
