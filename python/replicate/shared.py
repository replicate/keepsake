"""
The client for a simple RPC system to call Go.

Each request is a subprocess, it passes the request via stdin and receives the response via stdout.
It's like CGI RPC!

The server is in go/shared/main.go
"""

import base64
import json
import os
import subprocess
import sys


SHARED_BINARY = os.path.join(os.path.dirname(__file__), "bin/replicate-shared")


class InternalSharedError(Exception):
    pass


class SharedError(Exception):
    """
    An error from Go.

    Go's jsonrpc implementation doesn't let us pass error codes, so we use a predictable
    prefix to indicate error type.

    You should catch this exception and check the type to raise a Python error.
    """

    def __init__(self, message):
        super(SharedError, self).__init__(message)
        self.type = None
        self.message = message
        parts = str(self.message).split(":: ", 1)
        if len(parts) > 1:
            self.type = parts[0]
            self.message = parts[1]


# Here follows some bodges to let us pass binary data over json-rpc.
# Python -> Go is easy -- Go automatically decodes base64 into any []byte field.
# Go -> Python, any key called 'Data' we decode as base64.
#
# In an ideal world, we'd use a decent serialization format like protobuf or msgpack,
# but json-rpc is built into the Go stdlib and is trivial to use from Python. Either
# protobuf or msgpack would require much more implementation on the Python side, and
# we don't care about speed or size.


class SharedJSONEncoder(json.JSONEncoder):
    def default(self, o):
        if isinstance(o, bytes):
            return base64.b64encode(o).decode("ascii")
        return super(SharedJSONEncoder, self).default(o)


class SharedJSONDecoder(json.JSONDecoder):
    """
    A JSON decoder that decodes any key called 'data' as base64.
    """

    def __init__(self, *args, **kwargs):
        json.JSONDecoder.__init__(
            self, object_hook=self.object_hook_override, *args, **kwargs
        )

    def object_hook_override(self, obj):
        if "Data" in obj:
            obj["Data"] = base64.b64decode(obj["Data"])
        return obj


def call(method, **kwargs):
    request = {"version": "1.1", "method": method, "params": [kwargs]}
    result = subprocess.run(
        SHARED_BINARY,
        input=json.dumps(request, cls=SharedJSONEncoder).encode("utf-8"),
        stdout=subprocess.PIPE,
        check=True,
    )
    if result.stderr:
        print(result.stderr, file=sys.stderr)
    if not result.stdout:
        raise InternalSharedError("no response from shared binary")
    response = json.loads(result.stdout, cls=SharedJSONDecoder)
    if response.get("error"):
        raise SharedError(response["error"])
    return response["result"]
