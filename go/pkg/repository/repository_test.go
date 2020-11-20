package repository

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/replicate/replicate/go/pkg/files"
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
See the documentation for more details: https://replicate.ai/docs/reference/yaml`)), shim(SplitURL("foo://my-bucket")))
	require.Equal(t, shim(Scheme(""), "", "", fmt.Errorf(`Missing repository scheme.

Make sure your repository URL starts with either 'file://', 's3://', or 'gs://'.
See the documentation for more details: https://replicate.ai/docs/reference/yaml`)), shim(SplitURL("/foo/bar")))
}

func TestListOfFilesToPut(t *testing.T) {
	tmpDir, err := files.TempDir("repository-test")
	require.NoError(t, err)

	require.NoError(t, ioutil.WriteFile(filepath.Join(tmpDir, ".replicateignore"), []byte("ignoreme"), 0644))

	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "dir"), 0755))
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, ".git"), 0755))
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "ignoreme"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "my-venv", "bin"), 0755))
	require.NoError(t, ioutil.WriteFile(filepath.Join(tmpDir, "foo.txt"), []byte("foo"), 0644))
	require.NoError(t, ioutil.WriteFile(filepath.Join(tmpDir, "dir/bar.txt"), []byte("bar"), 0644))

	// test that venv is ignored
	require.NoError(t, ioutil.WriteFile(filepath.Join(tmpDir, "my-venv/pyvenv.cfg"), []byte("hello"), 0644))
	require.NoError(t, ioutil.WriteFile(filepath.Join(tmpDir, "my-venv/bin/activate"), []byte("world"), 0644))

	// test that .git is ignored
	require.NoError(t, ioutil.WriteFile(filepath.Join(tmpDir, ".git/baz.txt"), []byte("baz"), 0644))

	// test that .replicateignore is used
	require.NoError(t, ioutil.WriteFile(filepath.Join(tmpDir, "ignoreme/qux.txt"), []byte("qux"), 0644))

	filesToPut, err := getListOfFilesToPut(tmpDir, "")
	require.NoError(t, err)

	// erase .Info
	actual := []fileToPut{}
	for _, fp := range filesToPut {
		actual = append(actual, fileToPut{
			Source: fp.Source,
			Dest:   fp.Dest,
		})
	}

	expected := []fileToPut{{
		Source: filepath.Join(tmpDir, "foo.txt"),
		Dest:   "foo.txt",
	}, {
		Source: filepath.Join(tmpDir, "dir/bar.txt"),
		Dest:   "dir/bar.txt",
	}, {
		Source: filepath.Join(tmpDir, ".replicateignore"),
		Dest:   ".replicateignore",
	}}

	sort.Slice(actual, func(i, j int) bool { return actual[i].Source < actual[j].Source })
	sort.Slice(expected, func(i, j int) bool { return expected[i].Source < expected[j].Source })

	require.Equal(t, expected, actual)
}
