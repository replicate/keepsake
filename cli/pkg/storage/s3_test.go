package storage

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"replicate.ai/cli/pkg/hash"
)

func TestS3StorageGet(t *testing.T) {
	// It is possible to mock this stuff, but integration test is quick and easy
	// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/s3iface/
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		t.Skip("skipping S3 test because AWS_ACCESS_KEY_ID not set")
	}

	// Create a bucket
	bucketName := "replicate-test-" + hash.Random()[0:10]
	err := CreateS3Bucket("us-east-1", bucketName)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, DeleteS3Bucket("us-east-1", bucketName))
	}()
	// Even though CreateS3Bucket is supposed to wait until it exists, sometimes it doesn't
	time.Sleep(1 * time.Second)

	storage, err := NewS3Storage(bucketName, "root")
	require.NoError(t, err)

	require.NoError(t, storage.Put("some-file", []byte("hello")))

	data, err := storage.Get("some-file")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), data)

	_, err = storage.Get("does-not-exist")
	fmt.Println(err)
	require.IsType(t, &DoesNotExistError{}, err)
}
