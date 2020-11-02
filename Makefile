ENVIRONMENT := development

.PHONY: build
build:
	cd go && $(MAKE) build-all ENVIRONMENT=$(ENVIRONMENT)
	cd python && $(MAKE) build

.PHONY: develop
develop:
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
	git add go/Makefile python/setup.py web/.env
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
	cd go && gsutil cp -r release/ "gs://replicate-public/cli/$(VERSION)"
	cd go && gsutil cp -r release/ "gs://replicate-public/cli/latest"
	cd python && $(MAKE) build
	cd python && twine check dist/*
	cd python && twine upload dist/*

.PHONY: check-version-var
check-version-var:
	test $(VERSION)
