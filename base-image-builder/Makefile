NAME := replicate-image-builder
RELEASE_DIR := release
BINARY = $(RELEASE_DIR)/$(GOOS)/$(GOARCH)/$(NAME)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
MAIN := cmd/image-builder/main.go
INSTALL_PATH := /usr/local/bin/$(NAME)

all: build

build:
	@mkdir -p $(RELEASE_DIR)
	CGO_ENABLED=0 go build -o $(BINARY) $(MAIN)

install: build
	cp $(BINARY) $(INSTALL_PATH)
