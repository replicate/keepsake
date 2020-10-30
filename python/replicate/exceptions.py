class DoesNotExistError(Exception):
    pass


class UnknownRepositoryBackend(Exception):
    def __init__(self, scheme):
        super(UnknownRepositoryBackend, self).__init__(
            "Unknown repository backend: {}".format(scheme)
        )


class ConfigNotFoundError(Exception):
    def __init__(self, message):
        # TODO(andreas): global DOC_URL constant
        message += """

You must either create a replicate.yaml configuration file, or explicitly pass the arguments 'repository' and 'directory' to replicate.Project().

For more information, see https://replicate.ai/docs/reference/yaml"""
        super().__init__(message)
