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
	rm -rf build dist

# Build packages for the various platforms.
# setup.py handles converting these platform names into Go platforms.
#
# One trick worth noting with the build system: we're building "pure" Python
# libraries, that theoretically aren't tied to any Python runtime or platform.
# But, handily, wheel lets us override the platform with --plat-name even when it is a pure library:
# https://github.com/pypa/wheel/blob/730e3ae7427682435af9572126d2722adf7c438c/src/wheel/bdist_wheel.py#L237-L240
# I'm not sure we really should be doing this, because setting platform seems to be designed
# for non-pure modules, but what the heck, this seems to work?
# Anyway -- I'm writing this down here because if this stops working at some point in
# the future it's because wheel changed their logic around pure modules.
#
# For macosx, the version number indicates the _minimum_ version, so we just use an arbitrarily old one
# (the same one that numpy uses *shrug*)
# https://docs.python.org/3/distutils/apiref.html#distutils.util.get_platform
.PHONY: build
build: clean
	pip install wheel
	python setup.py bdist_wheel --plat-name manylinux1_i686
	python setup.py bdist_wheel --plat-name manylinux1_x86_64
	python setup.py bdist_wheel --plat-name macosx_10_9_x86_64

.PHONY: targets
targets:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'
