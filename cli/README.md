# Replicate command-line interface

## Install

    make install

## Run tests

    make test

You can pass additional arguments to `go test` with the `ARGS` variable. For example:

    make test ARGS="-run CheckpointNoRegistry"

## Release new version

Check your Git working directory is clean:

    $ git checkout master
    $ git pull
    $ git status

Then, bump the version:

    $ make bump-version

Then, run this script to create the release commit and tags and push them:

    $ make release
