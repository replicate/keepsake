# Replicate Python library

## Set up your local development environment

[Pyenv](https://github.com/pyenv/pyenv) makes it easy to switch between Python versions, and run tests against multiple Python version.

You can install pyenv with Homebrew (`brew install pyenv`) or with [pyenv-installer](https://github.com/pyenv/pyenv-installer).

Once it's installed, install Python 3.6, 3.7, and 3.8:

    $ pyenv install 3.6.9
    $ pyenv install 3.7.4
    $ pyenv install 3.8.0

Then make these versions available globally (or locally in the directory with `pyenv local` if you prefer).

    $ pyenv global 3.7.4 3.8.0 3.6.9

The first version is the one that is actually active, the other ones are made visible to tox (see below).

If you work in virtualenvs, [pyenv-virtualenv](https://github.com/pyenv/pyenv-virtualenv) neatly integrates virtualenv with pyenv.

Install [tox](https://tox.readthedocs.io/en/latest/) to run tests.

    $ pip install tox

## Run tests

    $ tox

[Tox](https://tox.readthedocs.io/en/latest/) creates virtual environments for various versions of Python, and runs the test suite against each environment.

### Run a single test

    $ tox -e py37 -- tests/unit/storage/test_s3_storage.py -k test_delete_exists

### Run integration tests

FIXME (bfirsh): document where to put keys

## Install for development

    $ python setup.py develop

This will make `import replicate` work anywhere on your machine, symlinked to this directory so it updates as you code.

## Build package

This will build an egg in `dist/`.

    $ python setup.py bdist_egg

## Release package

1.  Update the version string in these files:

    - `python/setup.py`
    - `cli/raw-assets/Dockerfile`

2.  Build the assets:

        $ make assets

3.  Commit this change ("Release Python x.y.z")
4.  Run this, which will build a tarball of the current version push it to `gs://replicate-python-dev/`:

        $ make release

## Format source code

Any contributions must be formatted with [Black](https://github.com/psf/black). The best thing is to set up your editor to automatically format code, but you can also do it manually by running:

    $ make fmt
