# type: ignore
# setuptools must be imported before distutils
import setuptools
from setuptools.command.develop import develop as _develop
from distutils.command.build_scripts import build_scripts as _build_scripts
from distutils.util import convert_path, get_platform
import os
from pathlib import Path
import re
import shutil
import ast

with open("../README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()


# Supported platforms we can build
#
# There doesn't seem to be a good list of platforms. The best I can find is the source code of distutils:
# https://github.com/python/cpython/blob/master/Lib/distutils/util.py
# Inside the packaging system, it then converts "-" to "_" to produce the platform name that goes in the filename
#
# See also:
# https://packaging.python.org/specifications/platform-compatibility-tags/
# https://www.python.org/dev/peps/pep-0425/
PLAT_NAME_TO_BINARY_PATH = {
    "linux_x86_64": "linux/amd64",
    # "linux" is the default if no --plat-name is passed, but it is not specific
    # enough for pypi, so we use manylinux for the released version
    "manylinux1_x86_64": "linux/amd64",
}


def plat_name_to_binary_path(plat_name):
    """
    Given a platform name, returns the path to the Go binaries to use.

    The platform name is the thing passed to `python setup.py bdist_wheel --plat-name ...`.
    """
    # These separators can be different things depending on where it comes from, so standardize
    plat_name = plat_name.replace("-", "_")
    plat_name = plat_name.replace(".", "_")
    # Simple ones
    if plat_name in PLAT_NAME_TO_BINARY_PATH:
        return PLAT_NAME_TO_BINARY_PATH[plat_name]
    # We need to do clever stuff for OS X, because it could be any version number
    if re.match(r"macosx_\d+_\d+_x86_64", plat_name):
        return "darwin/amd64"
    raise Exception("unsupported plat_name: " + plat_name)


def copy_binaries(plat_name):
    """
    Copy binaries for platform from ../go into current directory
    """
    this_dir = Path(__file__).resolve().parent
    binary_path = this_dir / "../go/release" / plat_name_to_binary_path(plat_name)
    (this_dir / "build/bin").mkdir(parents=True, exist_ok=True)
    (this_dir / "replicate/bin").mkdir(parents=True, exist_ok=True)
    shutil.copy(binary_path / "replicate", this_dir / "build/bin/replicate")
    shutil.copy(
        binary_path / "replicate-shared", this_dir / "replicate/bin/replicate-shared"
    )


# For stuff like `setup.py develop` and `setup.py install`, copy default binaries for
# this platform. When building wheels, these defaults will then be overridden.
copy_binaries(get_platform())


# wheel isn't always installed, so only override for when we're building packages
try:
    from wheel.bdist_wheel import bdist_wheel as _bdist_wheel

    # override bdist_wheel so we can copy binary into right place before wheel is created
    class bdist_wheel(_bdist_wheel):
        def run(self):
            copy_binaries(self.plat_name)
            _bdist_wheel.run(self)


except ImportError:
    bdist_wheel = None


# override build_scripts so we can install binaries
# The original expects them to be plain text:
# https://github.com/python/cpython/blob/master/Lib/distutils/command/build_scripts.py
class build_scripts(_build_scripts):
    def run(self):
        self.mkpath(self.build_dir)
        outfiles = []
        updated_files = []
        for script in self.scripts:
            script = convert_path(script)
            outfile = os.path.join(self.build_dir, os.path.basename(script))
            updated_files.append(outfile)
            outfiles.append(outfile)
            self.copy_file(script, outfile)
        return outfiles, updated_files


# don't install scripts for develop, because it expects them to be text
class develop(_develop):
    def install_egg_scripts(self, dist):
        pass


# read version from auto-generated version.py file
version = ast.parse(open("replicate/version.py").read()).body[0].value.s

# fmt: off
setuptools.setup(
    name="replicate",
    version=version,
    author_email="team@replicate.ai",
    description="Version control for machine learning",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://replicate.ai",
    license="Apache License 2.0",
    python_requires='>=3.6.0',
    packages=setuptools.find_packages(),
    package_data={'replicate': ['bin/replicate-shared']},
    cmdclass={
        'bdist_wheel': bdist_wheel,
        'build_scripts': build_scripts,
        'develop': develop,
    },
    scripts=["build/bin/replicate"],
)
# fmt: on
