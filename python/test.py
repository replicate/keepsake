from replicate.storage.disk_storage import DiskStorage

storage = DiskStorage("tmp")
print(storage.get("foof"))

