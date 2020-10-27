package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func shim(v ...interface{}) []interface{} {
	return v
}

// parallel of python/tests/unit/storage/test_storage.py
func TestSplitURL(t *testing.T) {
	require.Equal(t, shim(SchemeDisk, "", "/foo/bar", nil), shim(SplitURL("/foo/bar")))
	require.Equal(t, shim(SchemeDisk, "", "foo/bar", nil), shim(SplitURL("foo/bar")))
	require.Equal(t, shim(SchemeDisk, "", "/foo/bar", nil), shim(SplitURL("file:///foo/bar")))
	require.Equal(t, shim(SchemeDisk, "", "foo/bar", nil), shim(SplitURL("file://foo/bar")))

	require.Equal(t, shim(SchemeS3, "my-bucket", "", nil), shim(SplitURL("s3://my-bucket")))
	require.Equal(t, shim(SchemeS3, "my-bucket", "foo", nil), shim(SplitURL("s3://my-bucket/foo")))

	require.Equal(t, shim(SchemeGCS, "my-bucket", "", nil), shim(SplitURL("gs://my-bucket")))
	require.Equal(t, shim(SchemeGCS, "my-bucket", "foo", nil), shim(SplitURL("gs://my-bucket/foo")))

	require.Equal(t, shim(Scheme(""), "", "", fmt.Errorf("Unknown storage backend: foo")), shim(SplitURL("foo://my-bucket")))
}
