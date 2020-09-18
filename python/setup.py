# type: ignore
from distutils.command.build_scripts import build_scripts as _build_scripts
from distutils.util import convert_path
import os
from pathlib import Path
import setuptools
import shutil

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()


# The supported platforms we can build, passed to `python setup.py bdist_wheel --plat-name ...`
#
# For macosx, the version number indicates the _minimum_ version, so we just use an arbitrarily old one
# (the same one that numpy uses *shrug*)
# https://docs.python.org/3/distutils/apiref.html#distutils.util.get_platform
#
# There doesn't seem to be a good list of platforms. The best I can find is the source code of distutils:
# https://github.com/python/cpython/blob/master/Lib/distutils/util.py
# Inside the packaging system, it then converts "-" to "_" to produce the platform name that goes in the filename
#
# See also:
# https://packaging.python.org/specifications/platform-compatibility-tags/
# https://www.python.org/dev/peps/pep-0425/
PLAT_NAME_TO_BINARY = {
    "linux_i686": "linux/386/replicate",
    "linux_x86_64": "linux/amd64/replicate",
    # "linux" is the default if no --plat-name is passed, but it is not specific
    # enough for pypi, so define manylinux1 too
    "manylinux1_i686": "linux/386/replicate",
    "manylinux1_x86_64": "linux/amd64/replicate",
    "macosx_10_9_x86_64": "darwin/amd64/replicate",
}


# HACK: lots of setup.py commands rely on the script existing, so create a dummy one for when
# we're not building a package
this_dir = Path(__file__).resolve().parent
(this_dir / "build/bin").mkdir(parents=True, exist_ok=True)
(this_dir / "build/bin/replicate").touch()


def copy_binary(plat_name):
    """
    Copy binary for platform from ../cli into current directory
    """
    this_dir = Path(__file__).resolve().parent
    binary_path = this_dir / "../cli/release" / PLAT_NAME_TO_BINARY[plat_name]
    (this_dir / "build/bin").mkdir(parents=True, exist_ok=True)
    shutil.copy(binary_path, this_dir / "build/bin/replicate")


# wheel isn't always installed, so only override for when we're building packages
try:
    from wheel.bdist_wheel import bdist_wheel as _bdist_wheel

    # override bdist_wheel so we can copy binary into right place before wheel is created
    class bdist_wheel(_bdist_wheel):
        def run(self):
            copy_binary(self.plat_name)
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


# fmt: off
setuptools.setup(
    name="replicate",
    version="0.1.22",
    author="",
    author_email="",
    description="",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/replicate/replicate",
    python_requires='>=3.6.0',
    packages=setuptools.find_packages(),
    # TODO (bfirsh): maybe vendor all dependencies to make it not collide with other things you have installed
    # and break in weird ways?
    install_requires=[
        "boto3==1.12.32",
        "google-cloud-storage==1.23.0",
        "gitignore-parser==0.0.8",
        "pyyaml==5.3.1",
        "typing-extensions",
    ],
    extras_require={
        "test": [
            "moto==1.3.14",
            "mypy==0.782",
            "black==19.10b0",
            "pytest==5.4.3",
            "tox==3.14.1",
            "tensorflow==2.3.0",
        ],
    },
    setup_requires=[
        "cffi>=1.0.0",
    ],
    cffi_modules=["build_shared.py:ffibuilder"],
    cmdclass={
        'bdist_wheel': bdist_wheel,
        'build_scripts': build_scripts,
    },
    scripts=["build/bin/replicate"],
)
# fmt: on
