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

export GO111MODULE = on

.PHONY: default
default: build

.PHONY: all
all: clean build_all install

.PHONY: build
build:
	@mkdir -p $(RELEASE_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) $(MAIN)

.PHONY: build-all
build-all:
	@mkdir -p $(RELEASE_DIR)
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), \
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -v $(LDFLAGS) -o $(BINARY) $(MAIN); ))

install: build
	cp $(BINARY) $(INSTALL_PATH)

clean:
	rm -rf $(RELEASE_DIR)

test: check-fmt vet lint
	REPLICATE_NO_ANALYTICS=1 go test -v -timeout 1200s -parallel 5 ./... $(ARGS)


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
	@golint "" 2>/dev/null || (echo "installing golint..." && go get -u golang.org/x/lint/golint &>/dev/null)
	@golint ./... \
	| grep -v "should have comment.* or be unexported" \
	|| true
	@if [[ $$(wc -l /tmp/replicate-cli-lint-errors.txt | cut -f1 -d" ") > 0 ]]; then \
	    cat /tmp/replicate-cli-lint-errors.txt; \
	    exit 1; \
	else \
	    echo "golint successful"; \
	fi

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

.PHONY: docs
docs:
	REPLICATE_NO_ANALYTICS=1 go run ./cmd/replicate/main.go generate-docs --docs-folder web/docs
