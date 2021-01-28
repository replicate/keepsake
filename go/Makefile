SHELL := /bin/bash

VERSION := 0.3.3
ENVIRONMENT := development
PLATFORMS := darwin linux
ARCHITECTURES := amd64
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
MAIN := cmd/keepsake/main.go
SHARED_MAIN := cmd/keepsake-shared/main.go
NAME := keepsake
RELEASE_DIR := release
BINARY = $(RELEASE_DIR)/$(GOOS)/$(GOARCH)/$(NAME)
SHARED_BINARY = $(RELEASE_DIR)/$(GOOS)/$(GOARCH)/keepsake-shared
INSTALL_PATH := /usr/local/bin/$(NAME)

LDFLAGS := -ldflags "-X github.com/replicate/keepsake/go/pkg/global.Version=$(VERSION) -X github.com/replicate/keepsake/go/pkg/global.Environment=$(ENVIRONMENT) -w"

export GO111MODULE = on

# Local configuration
# This exports all the local make variables, but I think that's fine?
# https://stackoverflow.com/questions/44628206/how-to-load-and-export-variables-from-an-env-file-in-makefile
include .env.local
export

.env.local:
	touch .env.local


.PHONY: default
default: build

.PHONY: all
all: clean build_all install

.PHONY: build
build: clean
	@mkdir -p $(RELEASE_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) $(MAIN)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(SHARED_BINARY) $(SHARED_MAIN)

.PHONY: build-all
build-all:
	@mkdir -p $(RELEASE_DIR)
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), \
		GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) $(MAIN); \
		GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build $(LDFLAGS) -o $(SHARED_BINARY) $(SHARED_MAIN); \
	))

# install without sudo if the install path exists and is writeable,
# or if it doesn't exist and its directory is writeable
.PHONY: install
install: build
	if [[ (-f "$(INSTALL_PATH)" && -w "$(INSTALL_PATH)") || (! -f "$(INSTALL_PATH)" && -w $$(dirname "$(INSTALL_PATH)")) ]]; then \
	    cp $(BINARY) $(INSTALL_PATH); \
	else \
	    sudo cp $(BINARY) $(INSTALL_PATH); \
	fi

.PHONY: clean
clean:
	rm -rf $(RELEASE_DIR)

.PHONY: test
test: check-fmt vet lint
	go run gotest.tools/gotestsum -- -timeout 1200s -parallel 5 ./... $(ARGS)

.PHONY: test-external
test-external: check-fmt vet lint
	go run gotest.tools/gotestsum -- -tags=external -timeout 1200s -parallel 5 ./... $(ARGS)

BENCH := "."
.PHONY: benchmark
benchmark: build
	go test -bench=$(BENCH) -benchtime 10x -timeout 20m ./benchmark $(ARGS)


.PHONY: check-fmt
check-fmt:
	go run golang.org/x/tools/cmd/goimports --local keepsake.ai -d .
	@test -z $$(go run golang.org/x/tools/cmd/goimports --local keepsake.ai -l .)

.PHONY: vet
vet:
	go vet ./...

.PHONY: race
race:
	go build $(LDFLAGS) -race ./...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...

.PHONY: fmt
fmt:
	go run golang.org/x/tools/cmd/goimports --local keepsake.ai -w -d .

.PHONY: mod-tidy
mod-tidy:
	go mod tidy

.PHONY: check-new-version
check-new-version:
	[[ ! $$(gsutil ls "gs://replicate-public/cli/$(VERSION)") ]] 2>/dev/null

.PHONY: pre-commit
pre-commit: check-fmt vet lint

.PHONY: cover-profile
cover-profile: vet lint
	go test -timeout 1200s -coverprofile=/tmp/keepsake-cover-profile.out ./...
	go tool cover -html=/tmp/keepsake-cover-profile.out

.PHONY: strings
strings:
	find . -name "*.go" | grep -v _test.go | while read f; do echo $$f; grep -onE '"[^"]+"' $$f | grep -v "github.com/replicate/keepsake/go"; echo; done

# FIXME (bfirsh): get this working again
.PHONY: docs
docs:
	go run ./cmd/keepsake/main.go generate-docs --docs-folder ../web/pages/docs/reference/

# Check this is a valid Git commit
#
# This includes various things that we don't want to break the test suite when working
# day-to-day, but should be run on CI. E.g. checking generated files are committed,
# commit is signed off with Developer Certificate of Origin, etc.
#
# TODO (bfirsh): this should probably be moved to top-level Makefile
.PHONY: validate
validate: docs mod-tidy
	make/validate
