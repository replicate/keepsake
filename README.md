# Replicate

Version control for machine learning.

## Install

## Contributing

## Development environment

Run this to install the CLI and Python library locally for development:

    make develop

This will set up a symlink for the Python library. If you make changes to Go code, you will need to re-run this to compile and install.

## Build

This will build the Go CLI and the Python package:

    make build

The built Python packages are in `python/dist/`. These contain both the CLI and the Python library.

## Release

This will release both the CLI and Python package:

    make release VERSION=<version>

It pushes a new tag, which will trigger the "Release" Github action.

## Project structure

## VSCode development environment

If you use VSCode, we've included a workspace that will set up everything with the correct settings (formatting, autocomplete, and so on). Open it by running:

    code replicate.code-workspace
