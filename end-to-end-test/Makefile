ROOTDIR=../
VENV=$(ROOTDIR)venv/bin/

.PHONY: test
test:
	$(VENV)pytest -m "not external"

.PHONY: test-external
test-external:
	$(VENV)pytest
