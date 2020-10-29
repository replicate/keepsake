class DoesNotExistError(Exception):
    pass


class UnknownRepositoryBackend(Exception):
    def __init__(self, scheme):
        super(UnknownRepositoryBackend, self).__init__(
            "Unknown repository backend: {}".format(scheme)
        )
