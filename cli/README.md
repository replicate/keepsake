# Replicate command-line interface

## Install

    make install

## Run tests

    make test

You can pass additional arguments to `go test` with the `ARGS` variable. For example:

    make test ARGS="-run CheckpointNoRegistry"

## Run benchmarks

The benchmarks test the CLI against both the local disk and S3.

You'll need to configure some AWS credentials to run them locally. Create a file called `.env` in the current directory, and copy and paste your AWS credentials in:

    AWS_ACCESS_KEY_ID=
    AWS_SECRET_ACCESS_KEY=

Then, run the benchmarks:

    make benchmark

You can run specific benchmarks with the `BENCH` variable. For example:

    make benchmark BENCH="BenchmarkReplicateListOnDisk"
