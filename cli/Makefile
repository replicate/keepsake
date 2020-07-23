SHELL := /bin/bash

VERSION := 0.0.1
PLATFORMS := darwin linux windows
ARCHITECTURES := 386 amd64
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
MAIN := cmd/replicate/main.go
NAME := replicate
RELEASE_DIR := release
BINARY = $(RELEASE_DIR)/$(GOOS)/$(GOARCH)/$(NAME)
INSTALL_PATH := /usr/local/bin/$(NAME)

ASSETS_PACKAGE := assets
ASSETS_DIR := pkg/$(ASSETS_PACKAGE)
ASSETS_INPUT := raw-assets

LDFLAGS := -ldflags "-X replicate.ai/cli/pkg/global.Version=$(VERSION) -w"

export GO111MODULE = on

# Local configuration
# This exports all the local make variables, but I think that's fine?
# https://stackoverflow.com/questions/44628206/how-to-load-and-export-variables-from-an-env-file-in-makefile
include .env
export

.env:
	touch .env


.PHONY: default
default: build

.PHONY: all
all: clean build_all install

.PHONY: build
build: assets
	@mkdir -p $(RELEASE_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) $(MAIN)

.PHONY: build-all
build-all: assets
	@mkdir -p $(RELEASE_DIR)
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), \
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -v $(LDFLAGS) -o $(BINARY) $(MAIN); ))

.PHONY: install
install: build
	cp $(BINARY) $(INSTALL_PATH)

# TODO (bfirsh): this could be a real make statement to cache it. something like `pkg/assets/%.go: raw-assets/%`
# TODO(andreas): verify in CI that generated assets.go match raw-assets
.PHONY: assets
assets:
	@mkdir -p $(ASSETS_DIR)
	go run github.com/go-bindata/go-bindata/go-bindata -o $(ASSETS_DIR)/$(ASSETS_PACKAGE).go -pkg $(ASSETS_PACKAGE) -prefix $(ASSETS_INPUT)/ $(ASSETS_INPUT)

.PHONY: clean-assets
clean-assets:
	rm -rf $(ASSETS_DIR)

.PHONY: clean
clean: clean-assets
	rm -rf $(RELEASE_DIR)

test: check-fmt assets vet lint
	go run gotest.tools/gotestsum -- -timeout 1200s -parallel 5 ./... $(ARGS)


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

.PHONY: clean-deps
clean-deps:
	go mod tidy
	go mod verify

.PHONY: check-new-version
check-new-version:
	[[ ! $$(gsutil ls "gs://replicate-public/cli/$(VERSION)") ]] 2>/dev/null

.PHONY: pre-commit
pre-commit: check-fmt vet lint

.PHONY: bump-version
bump-version:
	script/bump-version

.PHONY: release
release:
	# TODO: check you haven't got a dirty checkout
	git checkout master
	git pull
	git commit -m "Bump to version $(VERSION)" Makefile
	git tag "v$(VERSION)"
	git push
	git push --tags

.PHONY: cover-profile
cover-profile: vet lint
	go test -timeout 1200s -coverprofile=/tmp/replicate-cover-profile.out ./...
	go tool cover -html=/tmp/replicate-cover-profile.out

.PHONY: strings
strings:
	find . -name "*.go" | grep -v _test.go | while read f; do echo $$f; grep -onE '"[^"]+"' $$f | grep -v "replicate.ai/cli"; echo; done

# FIXME (bfirsh): get this working again
.PHONY: docs
docs:
	go run ./cmd/replicate/main.go generate-docs --docs-folder web/docs
