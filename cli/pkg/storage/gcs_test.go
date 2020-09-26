package storage

import (
	"context"
	"io/ioutil"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"

	"replicate.ai/cli/pkg/hash"
)

// TODO: use Google's httpreplay library so this doesn't hit network
// https://godoc.org/cloud.google.com/go/httpreplay

// TODO: skip tests if you're not authenticated to Google Cloud. We should use `go test -short` to just run unit tests.

func createBucket(t *testing.T, client *storage.Client) (*storage.BucketHandle, string) {
	projectID, err := discoverProjectID()
	require.NoError(t, err)
	bucketName := "replicate-test-" + hash.Random()[0:10]
	bucket := client.Bucket(bucketName)
	err = bucket.Create(context.Background(), projectID, nil)
	require.NoError(t, err)
	return bucket, bucketName
}

func deleteBucket(t *testing.T, bucket *storage.BucketHandle) {
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

func TestGCSStorageGet(t *testing.T) {
	client, err := storage.NewClient(context.TODO())
	require.NoError(t, err)
	bucket, bucketName := createBucket(t, client)
	t.Cleanup(func() { deleteBucket(t, bucket) })
	createObject(t, bucket, "foo.txt", []byte("hello"))

	storage, err := NewGCSStorage(bucketName, "")
	require.NoError(t, err)
	data, err := storage.Get("foo.txt")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), data)
}

func TestGCSStoragePut(t *testing.T) {
	client, err := storage.NewClient(context.TODO())
	require.NoError(t, err)
	bucket, bucketName := createBucket(t, client)
	t.Cleanup(func() { deleteBucket(t, bucket) })

	storage, err := NewGCSStorage(bucketName, "")
	require.NoError(t, err)
	err = storage.Put("foo.txt", []byte("hello"))
	require.NoError(t, err)

	require.Equal(t, []byte("hello"), readObject(t, bucket, "foo.txt"))
}
