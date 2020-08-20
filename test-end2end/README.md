# End to end tests

## Install

Install the dependencies for the tests to run:

    pip install -r requirements.txt

When you run the end to end tests, they will install the CLI into `/usr/local/bin` and install the Python library globally.

## Running

To just run a simple local end to end test in development, run:

```
pytest -m fast
```

To run the full test suite:

```
pytest -s --aws-access-key-id=$CI_AWS_ACCESS_KEY_ID --aws-secret-access-key=$CI_AWS_SECRET_ACCESS_KEY --ssh-private-key=$CI_SSH_PRIVATE_KEY
```

The full test suite spins up EC2 instances and S3 buckets, so you need to pass it AWS credentials:

- `$CI_AWS_ACCESS_KEY_ID` and `$CI_SSH_PRIVATE_KEY` are AWS credentials
- `$CI_SSH_PRIVATE_KEY` is a path to a copy of the CI private key

FIXME (bfirsh): document setting up GCS backend
