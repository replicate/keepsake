# End to end tests

To run:

    $ (cd .. && make install-test-dependencies develop)
    $ make test

There are also some additional tests that hit Google Cloud and AWS. You first need to be signed into the `gcloud` and `aws` CLIs, and using test project/account. Then, run:

    make test-external
