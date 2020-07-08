# Replicate Python library

## Run tests

    $ python -m unittest

## Install for development

    $ python setup.py develop

This will make `import replicate` work anywhere on your machine, symlinked to this directory so it updates as you code.

## Build package

This will build an egg in `dist/`.

    $ python setup.py bdist_egg

## Format source code

Any contributions must be formatted with [Black](https://github.com/psf/black). The best thing is to set up your editor to automatically format code, but you can also do it manually by running:

    $ make format
