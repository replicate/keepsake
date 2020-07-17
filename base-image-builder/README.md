# Replicate image builder

Build Replicate base images.

We build two tiers of base images:
* Images without framework-specific packages (torch/tensorflow)
* Images with framework-specific packages. These inherit from the former

Both types are built with cpu and gpu versions, where gpu versions are built with various cuda/cudnn combinations.

For example, `us.gcr.io/replicate/base-ubuntu18.04-python3.8-cuda10.2-cudnn7-pytorch1.5.0:0.2` is a Ubuntu image with CUDA 10.2 and CuDNN 7, with Python 3.8 and PyTorch 1.5.0. The version tag `0.2` is an internal Replicate version so we can update images if we have to.

Base images are built manually by compiling and running this project. Images are built in parallel on Cloud Build, it takes around 3 hours (but is idempotent if it fails).

`cli/pkg/baseimages/baseimages.go` has a bunch of carefully copy-pasted compatibility tables. To avoid shooting ourselves in the foot, almost everything is type safe.

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
