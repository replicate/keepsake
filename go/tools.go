// +build tools

// https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md

package tools

import (
	_ "github.com/go-bindata/go-bindata/go-bindata"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "golang.org/x/tools/cmd/goimports"
	_ "gotest.tools/gotestsum"
)
