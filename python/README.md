# Replicate Python library

## Development environment

See `README.md` in the parent directory for instructions.

## Run tests

    $ (cd .. && make install-test-dependencies develop)
    $ make test

### Run a single test

To test a file:

    $ pytest tests/test_config.py

Or, an individual test:

    $ pytest tests/test_config.py -k test_validate

## Format source code

Any contributions must be formatted with [Black](https://github.com/psf/black). The best thing is to set up your editor to automatically format code, but you can also do it manually by running:

    $ make fmt

## Vendoring libraries

We vendor the few Python libraries we depend on to avoid dependency hell. We use [vendoring](https://pypi.org/project/vendoring/), the same tool used by pip.

The vendored packages are defined in `replicate/_vendor/vendor.txt`. If you add/change anything in there, run this to update the vendored libraries:

    make vendor
