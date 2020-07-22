.PHONY: pre-commit
pre-commit: lint

.PHONY: fmt
fmt:
	black .

.PHONY: test
test:
	tox

.PHONY: lint
lint:
	mypy .
	black --check .

.PHONY: clean
clean:
	rm -r dist

.PHONY: build
build: clean
	python setup.py sdist

.PHONY: release
release: build
	gsutil cp dist/*.tar.gz gs://replicate-python-dev/

.PHONY: targets
targets:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'
