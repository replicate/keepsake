import setuptools

with open("README.md", "r") as fh:
    long_description = fh.read()

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
        "boto3-stubs[essential]==1.14.15.0"
    ],
)
