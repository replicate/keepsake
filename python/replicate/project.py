import os

def get_project_dir():
    """
    Returns the directory of the current project.
    """
    # TODO (bfirsh): this currently is really simple and assumes you run your script in same dir as replicate.yaml.
    # But, this fails if you run your script from any other directory for whatever reason.
    # A better solution would be to go up the call stack until we find something non-replicate, figure out what directory
    # it is in, then search up the directory tree until we find replicate.yaml. That way replicate.yaml always works
    # regardless of working directory.
    return os.getcwd()
