// +build external

package repository

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/require"

	"github.com/replicate/replicate/go/pkg/files"
	"github.com/replicate/replicate/go/pkg/hash"
)

// It is possible to mock this stuff, but integration test is quick and easy
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/s3iface/
// TODO: perhaps use Google's httpreplay library so this doesn't hit network
// https://godoc.org/cloud.google.com/go/httpreplay

func TestS3RepositoryGet(t *testing.T) {
	bucketName, _ := createS3Bucket(t)
	t.Cleanup(func() { deleteS3Bucket(t, bucketName) })

	repository, err := NewS3Repository(bucketName, "root")
	require.NoError(t, err)

	require.NoError(t, repository.Put("some-file", []byte("hello")))

	data, err := repository.Get("some-file")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), data)

	_, err = repository.Get("does-not-exist")
	fmt.Println(err)
	require.IsType(t, &DoesNotExistError{}, err)
}

func TestS3GetPathTar(t *testing.T) {
	bucketName, _ := createS3Bucket(t)
	t.Cleanup(func() { deleteS3Bucket(t, bucketName) })

	repository, err := NewS3Repository(bucketName, "root")
	require.NoError(t, err)

	tmpDir, err := files.TempDir("test")
	require.NoError(t, err)
	err = repository.GetPathTar("does-not-exist.tar.gz", tmpDir)
	require.IsType(t, &DoesNotExistError{}, err)
}

func TestS3RepositoryPutPath(t *testing.T) {
	bucketName, svc := createS3Bucket(t)
	t.Cleanup(func() { deleteS3Bucket(t, bucketName) })

	tmpDir, err := files.TempDir("test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	err = os.Mkdir(filepath.Join(tmpDir, "somedir"), 0755)
	require.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(tmpDir, "somedir/foo.txt"), []byte("hello"), 0644)
	require.NoError(t, err)

	repository, err := NewS3Repository(bucketName, "")
	require.NoError(t, err)

	// Whole directory
	err = repository.PutPath(filepath.Join(tmpDir, "somedir"), "anotherdir")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), readS3Object(t, svc, bucketName, "anotherdir/foo.txt"))

	// Single file
	err = repository.PutPath(filepath.Join(tmpDir, "somedir/foo.txt"), "singlefile/foo.txt")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), readS3Object(t, svc, bucketName, "singlefile/foo.txt"))
}

func TestS3ListRecursive(t *testing.T) {
	bucketName, _ := createS3Bucket(t)
	t.Cleanup(func() { deleteS3Bucket(t, bucketName) })

	repository, err := NewS3Repository(bucketName, "")
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
	repository, err = NewS3Repository("replicate-test-"+hash.Random()[0:10], "")
	require.NoError(t, err)
	results = make(chan ListResult)
	go repository.ListRecursive(results, "checkpoints")
	require.Empty(t, <-results)
}

func createS3Bucket(t *testing.T) (string, *s3.S3) {
	bucketName := "replicate-test-" + hash.Random()[0:10]
	err := CreateS3Bucket("us-east-1", bucketName)
	require.NoError(t, err)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	require.NoError(t, err)
	return bucketName, s3.New(sess)
}

func deleteS3Bucket(t *testing.T, bucketName string) {
	require.NoError(t, DeleteS3Bucket("us-east-1", bucketName))
}

func readS3Object(t *testing.T, svc *s3.S3, bucketName string, key string) []byte {
	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	require.NoError(t, err)
	body, err := ioutil.ReadAll(obj.Body)
	require.NoError(t, err)
	return body
}
