# type: ignore

import setuptools

with open("README.md", "r") as fh:
    long_description = fh.read()

# fmt: off
setuptools.setup(
    name="replicate",
    version="0.0.1",
    author="",
    author_email="",
    description="",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/replicate/replicate",
    packages=setuptools.find_packages(),
    # TODO (bfirsh): maybe vendor all dependencies to make it not collide with other things you have installed
    # and break in weird ways?
    install_requires=[
        "pyyaml==5.3.1",
        "boto3==1.14.15",
        "boto3-stubs[essential]==1.14.15.0",
        "typing-extensions",
    ],
    extras_require={
        "test": [
            "moto==1.3.14",
            "mypy==0.782",
            "black==19.10b0",
            "pytest==5.4.3",
            "tox==3.14.1",
        ],
    }
)
# fmt: on
