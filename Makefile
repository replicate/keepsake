.PHONY: release
release: check-version-var verify-clean-master bump-version
	git commit -m "Bump to version $(VERSION)" Makefile
	git tag "v$(VERSION)"
	git push
	git push --tags

.PHONY: verify-version
# quick and dirty
bump-version:
	sed -E -i '' "s/VERSION := .+/VERSION := $(VERSION)/" cli/Makefile
	sed -E -i '' 's/version=".+"/version="$(VERSION)"/' python/setup.py
	sed -E -i '' 's/version: ".+"/version: "$(VERSION)"/' web/docusaurus.config.js

.PHONY: verify-clean-master
verify-clean-master:
	git diff-index --quiet HEAD  # make sure git is clean
	git checkout master
	git pull origin master

.PHONY: release-python
manually-release-python:
	cd python && \
	$(MAKE) build && \
	twine check dist/* && \
	twine upload dist/*

.PHONY: release-cli
manually-release-cli:
	cd cli && \
	$(MAKE) build-all && \
	gsutil cp -r release/ "gs://replicate-public/cli/$VERSION" && \
	gsutil cp -r release/ "gs://replicate-public/cli/latest"

.PHONY: check-version-var
check-version-var:
	test $(VERSION)
