from typing import Hashable, List, Sequence
import os

from ._shared import ffi, lib

# https://github.com/go-enry/go-enry/blob/master/python/enry/utils.py


def go_str_to_py(go_str):
    str_len = go_str.n
    if str_len > 0:
        return ffi.unpack(go_str.p, go_str.n).decode()
    return ""


def py_str_to_go(py_str: str):
    str_bytes = py_str.encode()
    c_str = ffi.new("char[]", str_bytes)
    go_str = ffi.new("_GoString_ *", [c_str, len(str_bytes)])
    # returning c_str stops it from getting garbage collected
    return (go_str[0], c_str)


def go_bool_to_py(go_bool):
    return go_bool == 1


def go_bytes_to_py_bytes(go_bytes) -> bytes:
    char_arr = ffi.cast("char *", go_bytes.data)
    return ffi.string(char_arr)


def disk_storage_get(root, path):
    root = os.path.join(os.getcwd(), root)

    ret_slice = ffi.new("GoSlice *")
    lib.DiskStorageGet(py_str_to_go(root)[0], py_str_to_go(path)[0], ret_slice)
    return go_bytes_to_py_bytes(ret_slice)

