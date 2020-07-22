package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Storage struct {
	bucketName string
	svc        *s3.S3
}

func NewS3Storage(bucket string) (*S3Storage, error) {
	region, err := discoverBucketRegion(bucket)
	if err != nil {
		return nil, fmt.Errorf("Failed to discover AWS region for bucket %s, got error: %s", bucket, err)
	}

	s := &S3Storage{
		bucketName: bucket,
	}
	sess, err := session.NewSession(&aws.Config{
		Region:                        aws.String(region),
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to S3, got error: %s", err)
	}
	s.svc = s3.New(sess)

	return s, nil
}

// Get data at path
func (s *S3Storage) Get(path string) ([]byte, error) {
	obj, err := s.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to read s3://%s/%s, got error: %s", s.bucketName, path, err)
	}
	body, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read body from s3://%s/%s, got error: %s", s.bucketName, path, err)
	}
	return body, nil
}

// Put data at path
func (s *S3Storage) Put(path string, data []byte) error {
	// TODO
	return nil
}

func (s *S3Storage) MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string) {
	s.listRecursive(results, folder, func(key string) bool {
		return filepath.Base(key) == filename
	})
}

func (s *S3Storage) listRecursive(results chan<- ListResult, folder string, filter func(string) bool) {
	// prefixes must end with / and must not end with /
	if !strings.HasSuffix(folder, "/") {
		folder += "/"
	}
	folder = strings.TrimPrefix(folder, "/")

	err := s.svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket:  aws.String(s.bucketName),
		Prefix:  aws.String(folder),
		MaxKeys: aws.Int64(1000),
	}, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, value := range page.Contents {
			key := *value.Key
			if filter(key) {
				results <- ListResult{Path: key}
			}
		}
		return lastPage
	})
	if err != nil {
		results <- ListResult{Error: fmt.Errorf("Failed to list objects in s3://%s, got error: %s", s.bucketName, err)}
	}
	close(results)
}

func discoverBucketRegion(bucket string) (string, error) {
	sess := session.Must(session.NewSession(&aws.Config{}))

	ctx := context.Background()
	region, err := s3manager.GetBucketRegion(ctx, sess, bucket, "eu-west-1")
	return region, err
}
