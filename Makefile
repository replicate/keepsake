ENVIRONMENT := development

.PHONY: build
build:
	cd cli && $(MAKE) build-all ENVIRONMENT=$(ENVIRONMENT)
	cd python && $(MAKE) build

.PHONY: develop
develop:
	cd cli && $(MAKE) build
	cd cli && $(MAKE) install
	cd python && python setup.py develop

.PHONY: install-test-dependencies
install-test-dependencies:
	pip install -r requirements-test.txt

.PHONY: test
test: install-test-dependencies develop
	cd cli && $(MAKE) test
	cd python && $(MAKE) test
	cd end-to-end-test && $(MAKE) test

.PHONY: test-external
test-external: install-test-dependencies develop
	cd cli && $(MAKE) test-external
	cd python && $(MAKE) test-external
	cd end-to-end-test && $(MAKE) test-external

.PHONY: release
release: check-version-var verify-clean-master bump-version
	git add cli/Makefile python/setup.py web/.env
	git commit -m "Bump to version $(VERSION)"
	git tag "v$(VERSION)"
	git push
	git push --tags

.PHONY: verify-version
# quick and dirty
bump-version:
	sed -E -i '' "s/VERSION := .+/VERSION := $(VERSION)/" cli/Makefile
	sed -E -i '' 's/version=".+"/version="$(VERSION)"/' python/setup.py
	sed -E -i '' 's/NEXT_PUBLIC_VERSION=.+/NEXT_PUBLIC_VERSION=$(VERSION)/' web/.env

.PHONY: verify-clean-master
verify-clean-master:
	git diff-index --quiet HEAD  # make sure git is clean
	git checkout master
	git pull origin master

.PHONY: release-manual
release-manual: check-version-var verify-clean-master
	cd cli && $(MAKE) build-all ENVIRONMENT=production
	cd cli && gsutil cp -r release/ "gs://replicate-public/cli/$(VERSION)" 
	cd cli && gsutil cp -r release/ "gs://replicate-public/cli/latest"
	cd python && $(MAKE) build
	cd python && twine check dist/*
	cd python && twine upload dist/*

.PHONY: check-version-var
check-version-var:
	test $(VERSION)
