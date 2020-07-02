import os

class DoesNotExistError(Exception): pass

class DiskStorage(object):
    """
    Stores data on local filesystem
    """

    def __init__(self, root):
        self.root = root

    def get(self, path):
        """
        Get data at path
        """
        full_path = os.path.join(self.root, path)
        try:
            with open(full_path) as fh:
                return fh.read()
        except FileNotFoundError:
            raise DoesNotExistError("No such path: '{}'".format(full_path))

    def put(self, path, data):
        """
        Save data to file at path
        """
        full_path = os.path.join(self.root, path)
        os.makedirs(os.path.dirname(full_path), exist_ok=True)

        mode = "w"
        if isinstance(data, bytes):
            mode = "wb"
        with open(full_path, mode) as fh:
            fh.write(data)

    # this can live in parent Storage class when that exists
    def put_directory(self, path, dir_to_store):
        """
        Save directory to path
        """
        for current_directory, _, files in os.walk(dir_to_store):
            for filename in files:
                with open(os.path.join(current_directory, filename), 'rb') as fh:
                    data = fh.read()
                # Strip local path
                relative_path = os.path.join(os.path.relpath(current_directory, dir_to_store), filename)
                # Then, make it relative to path we want to store it in storage
                self.put(os.path.join(path, relative_path), data)

    def list(self, path):
        """
        List files at path
        """
        # This is not recursive, but S3-style APIs make it very efficient to do recursive lists, so we probably want to add that
        full_path = os.path.join(self.root, path)
        for filename in os.listdir(full_path):
            if os.path.isfile(os.path.join(full_path, filename)):
                yield {"name": filename, "type": "file"}
            else:
                yield {"name": filename, "type": "directory"}

    def delete(self, path):
        """
        Delete single file at path
        """
        full_path = os.path.join(self.root, path)
        try:
            os.unlink(full_path)
        except FileNotFoundError:
            raise DoesNotExistError("No such path: '{}'".format(full_path))

