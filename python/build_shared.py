from cffi import FFI
import pathlib

current_dir = pathlib.Path(__file__).resolve().parent
lib_dir = current_dir.parent / "cli/release/shared/"
lib_header = lib_dir / "libreplicate.h"

ffibuilder = FFI()


def header_to_cdef(header_path):
    result = []
    skip = False
    with open(header_path) as fh:
        for line in fh:
            # skip cplusplus
            if line.startswith("#endif"):
                skip = False
                continue
            if skip:
                continue
            if line.startswith("#ifdef __cplusplus"):
                skip = True
                continue
            # ignore directives
            if line.startswith("#"):
                continue
            # unsupported things
            if "__SIZE_TYPE__" in line:
                continue
            if line.startswith("typedef char"):
                continue
            result.append(line)
    return "\n".join(result)


ffibuilder.cdef(header_to_cdef(lib_header))


ffibuilder.set_source(
    "replicate._shared",
    f'#include "{lib_header.absolute()}"',
    libraries=["replicate"],
    library_dirs=[str(lib_dir.absolute())],
    # https://stackoverflow.com/questions/58423739/go-static-library-on-qt5-undefined-symbols-for-architecture-x86-64
    extra_link_args=["-framework", "Security"],
)

if __name__ == "__main__":
    ffibuilder.compile(verbose=True)
