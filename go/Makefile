SHELL := /bin/bash

VERSION := 0.1.41
ENVIRONMENT := development
PLATFORMS := darwin linux windows
ARCHITECTURES := 386 amd64
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
MAIN := cmd/replicate/main.go
SHARED_MAIN := cmd/replicate-shared/main.go
NAME := replicate
RELEASE_DIR := release
BINARY = $(RELEASE_DIR)/$(GOOS)/$(GOARCH)/$(NAME)
SHARED_BINARY = $(RELEASE_DIR)/$(GOOS)/$(GOARCH)/replicate-shared
INSTALL_PATH := /usr/local/bin/$(NAME)

LDFLAGS := -ldflags "-X github.com/replicate/replicate/go/pkg/global.Version=$(VERSION) -X github.com/replicate/replicate/go/pkg/global.Environment=$(ENVIRONMENT) -w"

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

.PHONY: install
install: build
	cp $(BINARY) $(INSTALL_PATH)

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


.PHONY:
install-goimports:
	@goimports -h 2>&1 | grep -q usage || (go get golang.org/x/tools/cmd/goimports && echo "installed goimports")

.PHONY: check-fmt
check-fmt: install-goimports
	goimports --local replicate.ai -d .
	@test -z $$(goimports --local replicate.ai -l .)

.PHONY: vet
vet:
	go vet ./...

.PHONY: race
race:
	go build $(LDFLAGS) -race ./...

.PHONY: lint
lint:
	@golangci-lint help &>/dev/null || (echo "installing golangci-lint..." && go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0 &>/dev/null)
	golangci-lint run ./...

.PHONY: fmt
fmt: install-goimports
	goimports --local replicate.ai -w -d .

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
	go test -timeout 1200s -coverprofile=/tmp/replicate-cover-profile.out ./...
	go tool cover -html=/tmp/replicate-cover-profile.out

.PHONY: strings
strings:
	find . -name "*.go" | grep -v _test.go | while read f; do echo $$f; grep -onE '"[^"]+"' $$f | grep -v "github.com/replicate/replicate/go"; echo; done

# FIXME (bfirsh): get this working again
.PHONY: docs
docs:
	go run ./cmd/replicate/main.go generate-docs --docs-folder ../web/pages/docs/reference/

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
