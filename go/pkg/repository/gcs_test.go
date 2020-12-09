// +build external

package repository

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"

	"github.com/replicate/replicate/go/pkg/errors"
	"github.com/replicate/replicate/go/pkg/files"
	"github.com/replicate/replicate/go/pkg/hash"
)

// TODO: use Google's httpreplay library so this doesn't hit network
// https://godoc.org/cloud.google.com/go/httpreplay

// TODO: skip tests if you're not authenticated to Google Cloud. We should use `go test -short` to just run unit tests.

func createGCSBucket(t *testing.T, client *storage.Client) (*storage.BucketHandle, string) {
	projectID, err := discoverProjectID()
	require.NoError(t, err)
	bucketName := "replicate-test-" + hash.Random()[0:10]
	bucket := client.Bucket(bucketName)
	err = bucket.Create(context.Background(), projectID, nil)
	require.NoError(t, err)
	return bucket, bucketName
}

func deleteGCSBucket(t *testing.T, bucket *storage.BucketHandle) {
	it := bucket.Objects(context.Background(), &storage.Query{
		Prefix: "",
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		require.NoError(t, bucket.Object(attrs.Name).Delete(context.Background()))
	}
	require.NoError(t, bucket.Delete(context.Background()))
}

func clearGCSBucket(t *testing.T, bucket *storage.BucketHandle) {
	it := bucket.Objects(context.Background(), &storage.Query{
		Prefix: "",
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		require.NoError(t, bucket.Object(attrs.Name).Delete(context.Background()))
	}
}

func createObject(t *testing.T, bucket *storage.BucketHandle, key string, data []byte) {
	obj := bucket.Object(key)
	w := obj.NewWriter(context.Background())
	_, err := w.Write(data)
	require.NoError(t, err)
	require.NoError(t, w.Close())
}

func readObject(t *testing.T, bucket *storage.BucketHandle, key string) []byte {
	obj := bucket.Object(key)
	reader, err := obj.NewReader(context.Background())
	require.NoError(t, err)
	data, err := ioutil.ReadAll(reader)
	require.NoError(t, err)
	return data
}

// Run tests all in one go with one bucket because GCS rate limits bucket creation
func TestGCSRepository(t *testing.T) {
	client, err := storage.NewClient(context.TODO())
	require.NoError(t, err)
	bucket, bucketName := createGCSBucket(t, client)
	t.Cleanup(func() { deleteGCSBucket(t, bucket) })

	t.Run("Get", func(t *testing.T) {
		createObject(t, bucket, "foo.txt", []byte("hello"))
		repository, err := NewGCSRepository(bucketName, "")
		require.NoError(t, err)
		data, err := repository.Get("foo.txt")
		require.NoError(t, err)
		require.Equal(t, []byte("hello"), data)
	})

	clearGCSBucket(t, bucket)

	t.Run("GetPathTar", func(t *testing.T) {
		repository, err := NewGCSRepository(bucketName, "")
		require.NoError(t, err)

		tmpDir, err := files.TempDir("test")
		require.NoError(t, err)
		err = repository.GetPathTar("does-not-exist.tar.gz", tmpDir)
		require.True(t, errors.IsDoesNotExist(err))
	})

	clearGCSBucket(t, bucket)

	t.Run("Put", func(t *testing.T) {
		repository, err := NewGCSRepository(bucketName, "")
		require.NoError(t, err)
		err = repository.Put("foo.txt", []byte("hello"))
		require.NoError(t, err)

		require.Equal(t, []byte("hello"), readObject(t, bucket, "foo.txt"))
	})

	clearGCSBucket(t, bucket)

	t.Run("PutPath", func(t *testing.T) {
		tmpDir, err := files.TempDir("test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)
		err = os.Mkdir(path.Join(tmpDir, "somedir"), 0755)
		require.NoError(t, err)
		err = ioutil.WriteFile(path.Join(tmpDir, "somedir/foo.txt"), []byte("hello"), 0644)
		require.NoError(t, err)

		repository, err := NewGCSRepository(bucketName, "")
		require.NoError(t, err)

		// Whole directory
		err = repository.PutPath(filepath.Join(tmpDir, "somedir"), "anotherdir")
		require.NoError(t, err)
		require.Equal(t, []byte("hello"), readObject(t, bucket, "anotherdir/foo.txt"))

		// Single file
		err = repository.PutPath(filepath.Join(tmpDir, "somedir/foo.txt"), "singlefile/foo.txt")
		require.NoError(t, err)
		require.Equal(t, []byte("hello"), readObject(t, bucket, "singlefile/foo.txt"))
	})

	clearGCSBucket(t, bucket)

	t.Run("ListRecursive", func(t *testing.T) {
		repository, err := NewGCSRepository(bucketName, "")
		require.NoError(t, err)

		// Works with empty repository
		results := make(chan ListResult)
		go repository.ListRecursive(results, "checkpoints")
		require.Empty(t, <-results)

		// Lists stuff!
		require.NoError(t, repository.Put("checkpoints/abc123.json", []byte("yep")))
		require.NoError(t, repository.Put("experiments/def456.json", []byte("nope")))
		results = make(chan ListResult)
		go repository.ListRecursive(results, "checkpoints")
		require.Equal(t, ListResult{
			Path: "checkpoints/abc123.json",
			MD5:  []byte{0x93, 0x48, 0xae, 0x78, 0x51, 0xcf, 0x3b, 0xa7, 0x98, 0xd9, 0x56, 0x4e, 0xf3, 0x8, 0xec, 0x25},
		}, <-results)
		require.Empty(t, <-results)

		// Works with non-existent bucket
		repository, err = NewGCSRepository("replicate-test-"+hash.Random()[0:10], "")
		require.NoError(t, err)
		results = make(chan ListResult)
		go repository.ListRecursive(results, "checkpoints")
		require.Empty(t, <-results)
	})
}
