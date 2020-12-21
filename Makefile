ENVIRONMENT := development

OS := $(shell uname -s)

.PHONY: build
build: verify-dev-env
	cd go && $(MAKE) build-all ENVIRONMENT=$(ENVIRONMENT)
	cd python && $(MAKE) build

.PHONY: install
install: build
ifeq ($(OS),Linux)
	pip install python/dist/replicate-*-py3-none-manylinux1_x86_64.whl
else ifeq ($(OS),Darwin)
	pip install python/dist/replicate-*-py3-none-macosx_*.whl
else
	@echo Unknown OS: $(OS)
endif

.PHONY: develop
develop: verify-dev-env
	cd go && $(MAKE) build
	cd go && $(MAKE) install
	cd python && python setup.py develop

.PHONY: install-test-dependencies
install-test-dependencies:
	pip install -r requirements-test.txt

.PHONY: test
test: install-test-dependencies develop
	cd go && $(MAKE) test
	cd python && $(MAKE) test
	cd end-to-end-test && $(MAKE) test

.PHONY: test-external
test-external: install-test-dependencies develop
	cd go && $(MAKE) test-external
	cd python && $(MAKE) test-external
	cd end-to-end-test && $(MAKE) test-external

.PHONY: release
release: check-version-var verify-clean-main bump-version
	git add go/Makefile python/replicate/version.py web/.env
	git commit -m "Bump to version $(VERSION)"
	git tag "v$(VERSION)"
	git push git@github.com:replicate/replicate.git main
	git push git@github.com:replicate/replicate.git main --tags

.PHONY: verify-version
# quick and dirty
bump-version:
	sed -E -i '' "s/VERSION := .+/VERSION := $(VERSION)/" go/Makefile
	sed -E -i '' 's/version = ".+"/version = "$(VERSION)"/' python/replicate/version.py
	sed -E -i '' 's/NEXT_PUBLIC_VERSION=.+/NEXT_PUBLIC_VERSION=$(VERSION)/' web/.env

.PHONY: verify-clean-main
verify-clean-main:
	git diff-index --quiet HEAD  # make sure git is clean
	git checkout main
	git pull git@github.com:replicate/replicate.git main

.PHONY: release-manual
release-manual: check-version-var verify-clean-main
	cd go && $(MAKE) build-all ENVIRONMENT=production
	cd python && $(MAKE) build
	cd python && twine check dist/*
	cd python && twine upload dist/*

.PHONY: check-version-var
check-version-var:
	test $(VERSION)

.PHONY: verify-dev-env
verify-dev-env: verify-go-version verify-python-version

.PHONY: verify-go-version
verify-go-version:
	@./makefile-scripts/verify-go-version.sh

.PHONY: verify-python-version
verify-python-version:
	@./makefile-scripts/verify-python-version.sh
