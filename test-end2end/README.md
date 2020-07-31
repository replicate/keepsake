# End to end tests

These tests actually spin up EC2 instances and S3 buckets, so you need to pass it AWS credentials.

To run the tests:

```
pip install -r requirements.txt
pytest -s --aws-access-key-id=$CI_AWS_ACCESS_KEY_ID --aws-secret-access-key=$CI_AWS_SECRET_ACCESS_KEY --ssh-private-key=$CI_SSH_PRIVATE_KEY
```

where
* `$CI_AWS_ACCESS_KEY_ID` and `$CI_SSH_PRIVATE_KEY` are AWS credentials
* `$CI_SSH_PRIVATE_KEY` is a path to a copy of the CI private key

FIXME (bfirsh): document setting up GCS backend
