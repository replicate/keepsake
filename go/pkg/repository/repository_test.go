package repository

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func shim(v ...interface{}) []interface{} {
	return v
}

// parallel of python/tests/unit/repository/test_repository.py
func TestSplitURL(t *testing.T) {
	require.Equal(t, shim(SchemeDisk, "", "/foo/bar", nil), shim(SplitURL("file:///foo/bar")))
	require.Equal(t, shim(SchemeDisk, "", "foo/bar", nil), shim(SplitURL("file://foo/bar")))

	require.Equal(t, shim(SchemeS3, "my-bucket", "", nil), shim(SplitURL("s3://my-bucket")))
	require.Equal(t, shim(SchemeS3, "my-bucket", "foo", nil), shim(SplitURL("s3://my-bucket/foo")))

	require.Equal(t, shim(SchemeGCS, "my-bucket", "", nil), shim(SplitURL("gs://my-bucket")))
	require.Equal(t, shim(SchemeGCS, "my-bucket", "foo", nil), shim(SplitURL("gs://my-bucket/foo")))

	require.Equal(t, shim(Scheme(""), "", "", fmt.Errorf(`Unknown repository scheme: foo.

Make sure your repository URL starts with either 'file://', 's3://', or 'gs://'.
See the docuemntation for more details: https://replicate.ai/docs/reference/yaml`)), shim(SplitURL("foo://my-bucket")))
	require.Equal(t, shim(Scheme(""), "", "", fmt.Errorf(`Missing repository scheme.

Make sure your repository URL starts with either 'file://', 's3://', or 'gs://'.
See the docuemntation for more details: https://replicate.ai/docs/reference/yaml`)), shim(SplitURL("/foo/bar")))
}
