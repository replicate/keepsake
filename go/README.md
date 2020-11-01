# Replicate command-line interface

## Build / install

See instructions in `README.md` in the parent directory.

## Run tests

    make test

You can pass additional arguments to `go test` with the `ARGS` variable. For example:

    make test ARGS="-run CheckpointNoRegistry"

## Run benchmarks

The benchmarks test the CLI against both the local disk and S3:

    make benchmark

You can run specific benchmarks with the `BENCH` variable. For example:

    make benchmark BENCH="BenchmarkReplicateListOnDisk"
