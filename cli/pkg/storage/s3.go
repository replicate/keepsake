package storage

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"replicate.ai/cli/pkg/console"
)

type S3Storage struct {
	bucketName string
	sess       *session.Session
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
	s.sess, err = session.NewSession(&aws.Config{
		Region:                        aws.String(region),
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to S3, got error: %s", err)
	}
	s.svc = s3.New(s.sess)

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
	uploader := s3manager.NewUploader(s.sess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("Unable to upload %q to %q: %w", path, s.bucketName, err)
	}
	return nil
}

func (s *S3Storage) PutDirectory(localPath string, storagePath string) error {
	files, err := putDirectoryFiles(localPath, storagePath)
	if err != nil {
		return err
	}

	maxWorkers := int64(128)

	group, ctx := errgroup.WithContext(context.Background())
	group.Go(func() error {
		sem := semaphore.NewWeighted(maxWorkers)

		for _, file := range files {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}

			group.Go(func() error {
				defer sem.Release(1)

				data, err := ioutil.ReadFile(file.Source)
				if err != nil {
					return err
				}

				uploader := s3manager.NewUploader(s.sess)
				_, err = uploader.Upload(&s3manager.UploadInput{
					Bucket: aws.String(s.bucketName),
					Key:    aws.String(file.Dest),
					Body:   bytes.NewReader(data),
				})
				return err
			})
		}
		return nil
	})

	return group.Wait()
}

// GetDirectory recursively copies storageDir to localDir
func (s *S3Storage) GetDirectory(storageDir string, localDir string) error {
	iter := new(s3manager.DownloadObjectsIterator)
	files := []*os.File{}
	defer func() {
		for _, f := range files {
			if err := f.Close(); err != nil {
				console.Warn("Failed to close file %s", f.Name())
			}
		}
	}()

	keys := []*string{}
	err := s.svc.ListObjectsV2PagesWithContext(aws.BackgroundContext(), &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(storageDir),
	}, func(output *s3.ListObjectsV2Output, last bool) bool {
		for _, object := range output.Contents {
			keys = append(keys, object.Key)
		}
		return true
	})
	if err != nil {
		return fmt.Errorf("Failed to list objects in s3://%s/%s, got error: %w", s.bucketName, storageDir, err)
	}

	for _, key := range keys {
		// skip replicate-metadata
		if path.Base(*key) == "replicate-metadata.json" {
			continue
		}

		relPath, err := filepath.Rel(storageDir, *key)
		if err != nil {
			return fmt.Errorf("Failed to determine directory of %s relative to %s, got error: %w", *key, storageDir, err)
		}
		localPath := filepath.Join(localDir, relPath)
		localDir := filepath.Dir(localPath)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return fmt.Errorf("Failed to create directory %s, got error: %w", localDir, err)
		}

		f, err := os.Create(localPath)
		if err != nil {
			return fmt.Errorf("Failed to create file %s, got error: %w", localPath, err)
		}

		console.Debug("Downloading %s to %s", *key, localPath)

		iter.Objects = append(iter.Objects, s3manager.BatchDownloadObject{
			Object: &s3.GetObjectInput{
				Bucket: aws.String(s.bucketName),
				Key:    key,
			},
			Writer: f,
		})
		files = append(files, f)
	}

	downloader := s3manager.NewDownloader(s.sess)
	if err := downloader.DownloadWithIterator(aws.BackgroundContext(), iter); err != nil {
		return fmt.Errorf("Failed to download s3://%s/%s to %s", s.bucketName, storageDir, localDir)
	}
	return nil
}

func (s *S3Storage) MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string) {
	s.listRecursive(results, folder, func(key string) bool {
		return filepath.Base(key) == filename
	})
}

func CreateS3Bucket(region, bucket string) (err error) {
	sess, err := session.NewSession(&aws.Config{
		Region:                        aws.String(region),
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("Failed to connect to S3, got error: %w", err)
	}
	svc := s3.New(sess)

	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return fmt.Errorf("Unable to create bucket %q, %w", bucket, err)
	}

	// Default max attempts is 20, but we hit this sometimes
	return svc.WaitUntilBucketExistsWithContext(aws.BackgroundContext(), &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}, request.WithWaiterMaxAttempts(50))
}

func DeleteS3Bucket(region, bucket string) (err error) {
	sess, err := session.NewSession(&aws.Config{
		Region:                        aws.String(region),
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("Failed to connect to S3, got error: %w", err)
	}
	svc := s3.New(sess)

	iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	})

	if err := s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter); err != nil {
		return fmt.Errorf("Unable to delete objects from bucket %q, %w", bucket, err)
	}
	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return fmt.Errorf("Unable to delete bucket %q, %w", bucket, err)
	}
	return nil
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
		return true
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
