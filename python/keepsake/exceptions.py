from . import constants


class DoesNotExist(Exception):
    pass


class ReadError(Exception):
    pass


class WriteError(Exception):
    pass


class RepositoryConfigurationError(Exception):
    pass


class IncompatibleRepositoryVersion(Exception):
    pass


class CorruptedRepositorySpec(Exception):
    pass


class ConfigNotFound(Exception):
    pass
