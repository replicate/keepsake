# Replicate image builder

Build Replicate base images.

## Installation

1. Clone this repo.
1. You _must_ place this repo as a sibling of replicate-cli, and the replicate-cli folder _must_ be named `cli` (or else you have to modify `go.mod`).
1. `make install`

## Usage

1. Do whatever updates you need to the base images in `cli/bindata-assets/baseimages-*` and `cli/pkg/baseimages`.
1. `replicate-image-builder --version=0.X`
1. Check the logs on [Google Cloud Build](https://console.cloud.google.com/cloud-build)
1. Update `DefaultVersion` in `cli/pkg/baseimages/baseimages.go`

In case any builds fail, you can just re-run `replicate-image-builder --version=0.X`, it's idempotent and won't re-build images that have already been built.
