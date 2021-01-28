# Contributing guide

## Where to start?

You can contribute to Keepsake in a number of ways, depending on your expertise, experience, and interest.

If you want to **dive into the code**, the next sections of this document describes the project structure, commit instructions, etc. We have compiled a list of ["Good first issues"](https://github.com/replicate/keepsake/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22), which are small and well-defined tasks for getting started. We also have a ["Help wanted"](https://github.com/replicate/keepsake/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22) tag that has more general issues.

When you pick up an issue, please post a comment on the Github issue so others know that it's being worked on.

If you have **thoughts about ML versioning**, we would really appreciate a review of [the roadmap](https://github.com/replicate/keepsake/projects/1). You can use the thumbs up/down feature to let us know if you think a roadmap item is useful or not. It's also great if you comment on roadmap issues that you have specific thoughts about. If something is missing on the roadmap, please open a new issue and we'll discuss it there on Github.

The roadmap is a work in progress, not a waterfall-y specification, and we want it to evolve based on the input from the community!

We also host regular "community meetings" where we discuss specific roadmap items, and ML versioning and replicability in general. We send out invites to the meeting in advance, in the [`#community-meeting` Discord channel](https://discord.gg/qabvVCFv). Anyone is welcome to join, and they tend to be interesting discussions with a mix of industry practitioners and academics.

But perhaps the most useful thing you can do is **use the tool**. Join the [Discord chat](https://discord.gg/QmzJApGjyE) and let us know if Keepsake is useful to you, and how it can be improved. Open issues on [Github](https://github.com/replicate/keepsake) if you find bugs or have ideas for enhancements.

## Project structure

There are three main parts to the codebase:

- `go/`: This contains the `keepsake` command-line interface. It also provides a shared library that the Python library uses in `go/pkg/shared/`. The shared library runs in a standalone GRPC server in a Python subprocess.
- `python/`: This is the `keepsake` Python library. The Python package also includes the `keepsake` Go command-line interface and a Go shared library.
- `proto/`: This defines the interface between the Go server and the Python client.

The Python library acts as a thin client on top of the Go GRPC server, and all the heavy lifting is done in Go.

The other parts are:

- `end-to-end-test/`: A test suite that runs the Python library and Go CLI together against real S3/GCS buckets.
- `web/`: https://keepsake.ai

## Making a contribution

### Signing your work

Each commit you contribute to Keepsake must be signed off. It certifies that you wrote the patch, or have the right to contribute it. It is called the [Developer Certificate of Origin](https://developercertificate.org/) and was originally developed for the Linux kernel.

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

To generate the Protobuf implementations you need to install the required Protobuf tools. This is documented in `proto/Makefile`. Once they're installed, simply run `make build` from the `proto` folder.

## Release

This will release both the CLI and Python package:

    make release VERSION=<version>

It pushes a new tag, which will trigger the "Release" Github action.

## VSCode development environment

If you use VSCode, we've included a workspace that will set up everything with the correct settings (formatting, autocomplete, and so on). Open it by running:

    code keepsake.code-workspace
