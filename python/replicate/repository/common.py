from urllib.parse import urlparse

from .repository_base import Repository
from .disk_repository import DiskRepository
from .gcs_repository import GCSRepository
from .s3_repository import S3Repository

from ..exceptions import UnknownRepositoryBackend


def repository_for_url(url: str) -> Repository:
    parsed_url = urlparse(url)

    if parsed_url.scheme == "" or parsed_url.scheme == "file":
        # don't use os.path.join() here because path starts with "/" and join will treat that as root URL
        return DiskRepository(root=parsed_url.netloc + parsed_url.path)
    elif parsed_url.scheme == "s3":
        # lazy import to speed up import replicate

        return S3Repository(bucket=parsed_url.netloc, root=parsed_url.path.lstrip("/"))
    elif parsed_url.scheme == "gs":

        return GCSRepository(bucket=parsed_url.netloc, root=parsed_url.path.lstrip("/"))
    else:
        raise UnknownRepositoryBackend(parsed_url.scheme)
