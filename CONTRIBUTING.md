# Contributing guide

## Project structure

There are two main parts to the codebase:

- `go/`: This contains the `replicate` command-line interface. It also provides a shared library that the Python library uses in `go/pkg/shared/`. This is called with subprocess and jsonrpc via stdout/in (it's like CGI RPC!).
- `python/`: This is the `replicate` Python library. The Python package also includes the `replicate` Go command-line interface and a Go shared library.

The main mechanism that is shared between these two parts is the storage mechanism – reading/saving files on Amazon S3 or Google Cloud Storage. By implementing this in Go, we don't have to add a bazillion dependencies to the Python project. All other abstractions are mostly duplicated across the two languages (repositories, experiments, checkpoints, etc), but this line might move over time.

The other parts are:

- `end-to-end-test/`: A test suite that runs the Python library and Go CLI together against real S3/GCS buckets.
- `web/`: https://replicate.ai

## Making a contribution

### Signing your work

Each commit you contribute to Replicate must be signed off. It certifies that you wrote the patch, or have the right to contribute it. It is called the [Developer Certificate of Origin](https://developercertificate.org/) and was originally developed for the Linux kernel.

If you can certify the following:

```
By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

Then add this line to each of your Git commit messages, with your name and email:

```
Signed-off-by: Joe Smith <joe.smith@email.com>
```

You can sign your commit automatically by passing the `-s` option to Git commit: `git commit -s -m "Reticulate splines"`

## Development environment

Run this to install the CLI and Python library locally for development:

    make develop

This will set up a symlink for the Python library. If you make changes to Go code, you will need to re-run this to compile and install.

## Test

Run this to run the test suite:

    make test

This will run the three test suites in the `go/`, `python/`, and `end-to-end-tests/` directories. You can also run `make test` in those directories to run the test suites individually, after running `make develop` in the root directory to install everything correctly.

There are also some additional tests that hit Google Cloud and AWS. You first need to be signed into the `gcloud` and `aws` CLIs, and using test project/account. Then, run:

    make test-external

## Build

This will build the CLI and the Python package:

    make build

The built Python packages are in `python/dist/`. These contain both the CLI and the Python library.

## Release

This will release both the CLI and Python package:

    make release VERSION=<version>

It pushes a new tag, which will trigger the "Release" Github action.

## VSCode development environment

If you use VSCode, we've included a workspace that will set up everything with the correct settings (formatting, autocomplete, and so on). Open it by running:

    code replicate.code-workspace
