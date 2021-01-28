#!/bin/bash -eu

INSTALL_MESSAGE="Please follow the instructions on https://golang.org/doc/install to install the latest version of Go."

GO_VERSION=$(go version 2>&1 || true)

if $(echo "$GO_VERSION" | grep -q -E 'go: (command )?not found'); then
    echo "ERROR: Go is not installed."
    echo
    echo "$INSTALL_MESSAGE"
    exit 1
fi

if $(echo "$GO_VERSION" | grep -q -E -v "go version"); then
    echo "ERROR: failed to determine go version, 'go version' returned:"
    echo "  $GO_VERSION"
    echo
    echo "$INSTALL_MESSAGE"
    exit 1
fi

GO_VERSION_NUMBER=$(echo "$GO_VERSION" | sed -E 's/^go version go([^ ]+) .+$/\1/')

if $(echo "$GO_VERSION_NUMBER" | grep -q -E -v '^1\.1[456]'); then
    echo "ERROR: Unsupported Go version: $GO_VERSION_NUMBER"
    echo "Keepsake requires Go >= 1.14"
    echo
    echo "$INSTALL_MESSAGE"
    exit 1
fi
