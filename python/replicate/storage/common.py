import re

from .storage_base import Storage
from .disk_storage import DiskStorage
from .s3_storage import S3Storage

from ..exceptions import UnknownStorageBackend


def storage_for_url(url: str) -> Storage:
    url_re = re.compile("^([^:]+)://(.+)$")
    match = url_re.match(url)
    if match is None:
        return DiskStorage(root=url)

    scheme, path = match.groups()

    if scheme == "s3":
        return S3Storage(bucket=path)
    if scheme == "file":
        return DiskStorage(root=path)
    else:
        raise UnknownStorageBackend(scheme)
