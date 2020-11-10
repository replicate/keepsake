from . import constants


class DoesNotExistError(Exception):
    pass


class UnknownRepositoryScheme(Exception):
    def __init__(self, scheme):
        if scheme == "":
            message = "Missing repository scheme"
        else:
            message = "Unknown repository scheme: {}".format(scheme)
        super().__init__(
            message
            + """.

Make sure your repository URL starts with either 'file://', 's3://', or 'gs://'.
See the documentation for more details: {}""".format(
                constants.YAML_REFERENCE_DOCS_URL
            )
        )


class ConfigNotFoundError(Exception):
    def __init__(self, message):
        message += """

You must either create a replicate.yaml configuration file, or explicitly pass the arguments 'repository' and 'directory' to replicate.Project().

For more information, see {}""".format(
            constants.YAML_REFERENCE_DOCS_URL
        )
        super().__init__(message)


class NewerRepositoryVersion(Exception):
    def __init__(self, repository_url):
        message = """The repository at {} is using a newer storage mechanism which is incompatible with your version of Replicate.

To upgrade, run:
pip install --upgrade replicate
""".format(
            repository_url
        )
        super().__init__(message)


class CorruptedProjectSpec(Exception):
    def __init__(self, path):
        message = """The project spec file at {} is corrupted.

You can manually edit it with the format {"version": VERSION},
where VERSION is an integer."""
        super().__init__(message)
