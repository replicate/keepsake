class DoesNotExistError(Exception):
    pass


class UnknownStorageBackend(Exception):
    def __init__(self, scheme):
        super(UnknownStorageBackend, self).__init__(
            "Unknown storage backend: {}".format(scheme)
        )
