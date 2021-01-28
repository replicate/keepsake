package repository

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/replicate/keepsake/go/pkg/errors"
	"github.com/replicate/keepsake/go/pkg/files"
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
See the documentation for more details: https://keepsake.ai/docs/reference/yaml`)), shim(SplitURL("foo://my-bucket")))
	require.Equal(t, shim(Scheme(""), "", "", fmt.Errorf(`Missing repository scheme.

Make sure your repository URL starts with either 'file://', 's3://', or 'gs://'.
See the documentation for more details: https://keepsake.ai/docs/reference/yaml`)), shim(SplitURL("/foo/bar")))
}

func TestListOfFilesToPut(t *testing.T) {
	tmpDir, err := files.TempDir("repository-test")
	require.NoError(t, err)

	require.NoError(t, ioutil.WriteFile(filepath.Join(tmpDir, ".keepsakeignore"), []byte("ignoreme"), 0644))

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

	// test that .keepsakeignore is used
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
		Source: filepath.Join(tmpDir, ".keepsakeignore"),
		Dest:   ".keepsakeignore",
	}}

	sort.Slice(actual, func(i, j int) bool { return actual[i].Source < actual[j].Source })
	sort.Slice(expected, func(i, j int) bool { return expected[i].Source < expected[j].Source })

	require.Equal(t, expected, actual)
}

func TestExtractTarItem(t *testing.T) {
	dir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create some temporary files
	fileDir := path.Join(dir, "files")
	err = os.MkdirAll(fileDir, os.ModePerm)
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(fileDir, "a.txt"), []byte("file a"), 0644)
	require.NoError(t, err)
	err = ioutil.WriteFile(path.Join(fileDir, "b.txt"), []byte("file b"), 0644)
	require.NoError(t, err)
	err = os.Mkdir(path.Join(fileDir, "c"), 0755)
	require.NoError(t, err)
	err = ioutil.WriteFile(path.Join(fileDir, "c/d.txt"), []byte("file d"), 0644)
	require.NoError(t, err)

	// Archive the sub-directory as a tarball in the repository
	// This should result in a tarball with the following directory tree:
	//
	// temp
	// |--c
	// |  |-- d.txt
	// |
	// |-- a.txt
	// |-- b.txt
	tarFile, err := os.Create(path.Join(dir, "temp.tar.gz"))
	require.NoError(t, err)
	defer tarFile.Close()

	err = putPathTar(fileDir, tarFile, "temp.tar.gz", "")
	require.NoError(t, err)

	// Create a temporary directory
	tmpDir, err := files.TempDir("test")
	require.NoError(t, err)

	// Extract just one of the two files from the repo dir.
	err = extractTarItem(path.Join(dir, "temp.tar.gz"), "a.txt", tmpDir)
	require.NoError(t, err)

	content, err := ioutil.ReadFile(path.Join(tmpDir, "a.txt"))
	require.NoError(t, err)
	require.Equal(t, []byte("file a"), content)

	// Extract an entire directory
	err = extractTarItem(path.Join(dir, "temp.tar.gz"), "c", tmpDir)
	require.NoError(t, err)

	content, err = ioutil.ReadFile(path.Join(tmpDir, "c/d.txt"))
	require.NoError(t, err)
	require.Equal(t, []byte("file d"), content)

	// Extract a file that does not exist
	err = extractTarItem(path.Join(dir, "temp.tar.gz"), "does-not-exist.txt", tmpDir)
	require.True(t, errors.IsDoesNotExist(err))
}

func TestCopyToTempDir(t *testing.T) {
	dir, err := files.TempDir("test")
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(dir, "foo"), []byte("foo"), 0644)
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(dir, "my/folder"), 0755)
	require.NoError(t, err)
	err = ioutil.WriteFile(path.Join(dir, "my/folder/bar"), []byte("bar"), 0644)
	require.NoError(t, err)

	// without includePath
	tempDir, err := CopyToTempDir(dir, ".")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	contents, err := ioutil.ReadFile(path.Join(tempDir, "foo"))
	require.NoError(t, err)
	require.Equal(t, "foo", string(contents))

	contents, err = ioutil.ReadFile(path.Join(tempDir, "my/folder/bar"))
	require.NoError(t, err)
	require.Equal(t, "bar", string(contents))

	// with directory includePath
	tempDir, err = CopyToTempDir(dir, "my")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	exists, err := files.FileExists(path.Join(tempDir, "foo"))
	require.NoError(t, err)
	require.False(t, exists)

	contents, err = ioutil.ReadFile(path.Join(tempDir, "my/folder/bar"))
	require.NoError(t, err)
	require.Equal(t, "bar", string(contents))

	// with file includePath
	tempDir, err = CopyToTempDir(dir, "my/folder/bar")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	exists, err = files.FileExists(path.Join(tempDir, "foo"))
	require.NoError(t, err)
	require.False(t, exists)

	contents, err = ioutil.ReadFile(path.Join(tempDir, "my/folder/bar"))
	require.NoError(t, err)
	require.Equal(t, "bar", string(contents))
}
