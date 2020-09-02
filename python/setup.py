# type: ignore

import setuptools

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

# fmt: off
setuptools.setup(
    name="replicate",
    version="0.1.15",
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
        "aiobotocore==1.0.7",
        "boto3-stubs[essential]==1.12.32.0",
        "boto3==1.12.32",
        "botocore==1.15.32",
        "gcloud-aio-storage==5.5.4",
        "google-cloud-storage==1.23.0",
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
        ],
    }
)
# fmt: on
