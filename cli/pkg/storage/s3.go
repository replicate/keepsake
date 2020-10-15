package storage

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/adjust/uniuri"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts"
	"golang.org/x/sync/errgroup"

	"replicate.ai/cli/pkg/concurrency"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/files"
	"replicate.ai/cli/pkg/settings"
)

type S3Storage struct {
	bucketName string
	root       string
	sess       *session.Session
	svc        *s3.S3
}

func NewS3Storage(bucket, root string) (*S3Storage, error) {
	region, err := getBucketRegionOrCreateBucket(bucket)
	if err != nil {
		return nil, err
	}

	s := &S3Storage{
		bucketName: bucket,
		root:       root,
	}
	s.sess, err = session.NewSession(&aws.Config{
		Region:                        aws.String(region),
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to S3: %s", err)
	}
	s.svc = s3.New(s.sess)

	return s, nil
}

func (s *S3Storage) RootURL() string {
	ret := "s3://" + s.bucketName
	if s.root != "" {
		ret += "/" + s.root
	}
	return ret
}

// Get data at path
func (s *S3Storage) Get(path string) ([]byte, error) {
	key := filepath.Join(s.root, path)
	obj, err := s.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey {
				return nil, &DoesNotExistError{msg: "Get: path does not exist: " + path}
			}
		}
		return nil, fmt.Errorf("Failed to read %s/%s: %s", s.RootURL(), path, err)
	}
	body, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read body from %s/%s: %s", s.RootURL(), path, err)
	}
	return body, nil
}

func (s *S3Storage) Delete(path string) error {
	console.Debug("Deleting %s/%s...", s.RootURL(), path)
	key := filepath.Join(s.root, path)
	iter := s3manager.NewDeleteListIterator(s.svc, &s3.ListObjectsInput{
		Bucket: &s.bucketName,
		Prefix: &key,
	})
	if err := s3manager.NewBatchDeleteWithClient(s.svc).Delete(aws.BackgroundContext(), iter); err != nil {
		return fmt.Errorf("Failed to delete %s/%s: %w", s.RootURL(), path, err)
	}
	return nil
}

// Put data at path
func (s *S3Storage) Put(path string, data []byte) error {
	key := filepath.Join(s.root, path)
	uploader := s3manager.NewUploader(s.sess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("Unable to upload to %s/%s: %w", s.RootURL(), path, err)
	}
	return nil
}

func (s *S3Storage) PutPath(localPath string, destPath string) error {
	files, err := getListOfFilesToPut(localPath, filepath.Join(s.root, destPath))
	if err != nil {
		return err
	}
	queue := concurrency.NewWorkerQueue(context.Background(), maxWorkers)

	for _, file := range files {
		// Variables used in closure
		file := file
		err := queue.Go(func() error {
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
		if err != nil {
			return err
		}
	}

	return queue.Wait()
}

func (s *S3Storage) PutPathTar(localPath, tarPath, includePath string) error {
	if !strings.HasSuffix(tarPath, ".tar.gz") {
		return fmt.Errorf("PutPathTar: tarPath must end with .tar.gz")
	}

	reader, writer := io.Pipe()

	// TODO: This doesn't cancel elegantly on error -- we should use the context returned here and check if it is done.
	errs, _ := errgroup.WithContext(context.TODO())

	errs.Go(func() error {
		if err := putPathTar(localPath, writer, filepath.Base(tarPath), includePath); err != nil {
			return err
		}
		return writer.Close()
	})
	errs.Go(func() error {
		key := filepath.Join(s.root, tarPath)
		uploader := s3manager.NewUploader(s.sess)
		_, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.bucketName),
			Key:    aws.String(key),
			Body:   reader,
		})
		return err
	})
	return errs.Wait()
}

// GetPath recursively copies storageDir to localDir
func (s *S3Storage) GetPath(remoteDir string, localDir string) error {
	prefix := filepath.Join(s.root, remoteDir)
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
		Prefix: aws.String(prefix),
	}, func(output *s3.ListObjectsV2Output, last bool) bool {
		for _, object := range output.Contents {
			keys = append(keys, object.Key)
		}
		return true
	})
	if err != nil {
		return fmt.Errorf("Failed to list objects in s3://%s/%s: %w", s.bucketName, prefix, err)
	}

	for _, key := range keys {
		relPath, err := filepath.Rel(prefix, *key)
		if err != nil {
			return fmt.Errorf("Failed to determine directory of %s relative to %s: %w", *key, prefix, err)
		}
		localPath := filepath.Join(localDir, relPath)
		localDir := filepath.Dir(localPath)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return fmt.Errorf("Failed to create directory %s: %w", localDir, err)
		}

		f, err := os.Create(localPath)
		if err != nil {
			return fmt.Errorf("Failed to create file %s: %w", localPath, err)
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
		return fmt.Errorf("Failed to download s3://%s/%s to %s", s.bucketName, prefix, localDir)
	}
	return nil
}

func (s *S3Storage) GetPathTar(tarPath, localPath string) error {
	// archiver doesn't let us use readers, so download to temporary file
	// TODO: make a better tar implementation
	tmpdir, err := files.TempDir("tar")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)
	tmptarball := filepath.Join(tmpdir, filepath.Base(tarPath))
	if err := s.GetPath(tarPath, tmptarball); err != nil {
		return err
	}
	return extractTar(tmptarball, localPath)
}

func (s *S3Storage) ListRecursive(results chan<- ListResult, dir string) {
	s.listRecursive(results, dir, func(_ string) bool { return true })
}

func (s *S3Storage) MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string) {
	s.listRecursive(results, folder, func(key string) bool {
		return filepath.Base(key) == filename
	})
}

// List files in a path non-recursively
func (s *S3Storage) List(dir string) ([]string, error) {
	results := []string{}
	prefix := filepath.Join(s.root, dir)

	// prefixes must end with / and must not end with /
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	prefix = strings.TrimPrefix(prefix, "/")

	err := s.svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket:    aws.String(s.bucketName),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int64(1000),
	}, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, value := range page.Contents {
			key := *value.Key
			if s.root != "" {
				key = strings.TrimPrefix(strings.TrimPrefix(key, s.root), "/")
			}
			results = append(results, key)
		}
		return true
	})
	return results, err
}

func (s *S3Storage) PrepareRunEnv() ([]string, error) {
	accessKey, err := s.getOrCreateIAMUser()
	if err != nil {
		return nil, err
	}
	return []string{
		"AWS_ACCESS_KEY_ID=" + *accessKey.AccessKeyId,
		"AWS_SECRET_ACCESS_KEY=" + *accessKey.SecretAccessKey,
		"AWS_DEFAULT_REGION=" + *s.sess.Config.Region,
	}, nil
}

func (s *S3Storage) getOrCreateIAMUser() (*iam.AccessKey, error) {
	secretName := "aws-iam-user-" + s.bucketName
	data, err := settings.GetSecret(secretName)
	if err != nil {
		return nil, err
	}
	if data != nil {
		var accessKey *iam.AccessKey
		err := json.Unmarshal(data, &accessKey)
		return accessKey, err
	}
	key, err := s.createIAMUser()
	if err != nil {
		return nil, err
	}
	data, err = json.Marshal(key)
	if err != nil {
		return nil, err
	}
	if err := settings.SetSecret(secretName, data); err != nil {
		return nil, err
	}
	return key, nil

}

func (s *S3Storage) createIAMUser() (*iam.AccessKey, error) {
	// AWS have a service for these sorts of access keys called "STS".
	// We're using regular IAM accounts instead because those have a max expiry for 36 hours.
	// Training jobs might run for much longer than that, so we need regular IAM users.

	console.Info("Creating an IAM account with limited permissions to access \"s3://%s\"...", s.bucketName)

	svc := iam.New(s.sess)
	// https://docs.aws.amazon.com/IAM/latest/APIReference/API_User.html
	maxBucketNameLength := 64 - len("replicate") - 7 - 2
	bucketName := s.bucketName
	if len(bucketName) > maxBucketNameLength {
		bucketName = bucketName[:maxBucketNameLength]
	}
	username := fmt.Sprintf("replicate-%s-%s", bucketName, strings.ToLower(uniuri.NewLen(7)))

	// Create user
	_, err := svc.CreateUser(&iam.CreateUserInput{
		UserName: aws.String(username),
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating IAM user: %w", err)
	}

	// Create policy
	policyDocument, err := s.iamPolicy()
	if err != nil {
		return nil, err
	}
	createPolicyResult, err := svc.CreatePolicy(&iam.CreatePolicyInput{
		PolicyDocument: aws.String(policyDocument),
		PolicyName:     aws.String(username),
	})
	if err != nil {
		return nil, err
	}

	// Attach policy to user
	_, err = svc.AttachUserPolicy(&iam.AttachUserPolicyInput{
		PolicyArn: aws.String(*createPolicyResult.Policy.Arn),
		UserName:  aws.String(username),
	})
	if err != nil {
		return nil, err
	}

	// Create access key
	result, err := svc.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(username),
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating IAM access key: %w", err)
	}

	// Wait for IAM user to exist
	// The IAM user exists now if you call the IAM API, but the access key hasn't yet propagated,
	// so we need to call an actual API call and check that it doesn't throw invalid access key *heavy sigh*
	console.Info("Waiting for IAM user to exist...")
	err = waitForOperation(context.TODO(), func() (bool, error) {
		sess, err := session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(*result.AccessKey.AccessKeyId, *result.AccessKey.SecretAccessKey, ""),
		})
		if err != nil {
			return false, err
		}

		_, err = sts.New(sess).GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err != nil {
			// This error doesn't seem to come out with a proper type
			if strings.Contains(err.Error(), "InvalidClientTokenId") {
				return false, nil
			}
			return false, err
		}
		return true, nil
	})
	return result.AccessKey, err
}

type policyDocument struct {
	Version   string
	Statement []statementEntry
}

type statementEntry struct {
	Effect   string
	Action   []string
	Resource []string
}

// Returns IAM policy JSON which allows full access to a single S3 bucket
func (s *S3Storage) iamPolicy() (string, error) {
	policy := policyDocument{
		Version: "2012-10-17",
		Statement: []statementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"s3:GetBucketLocation",
					"s3:ListBucket",
				},
				Resource: []string{"arn:aws:s3:::" + s.bucketName},
			},
			{
				Effect: "Allow",
				Action: []string{
					"s3:*",
				},
				Resource: []string{
					"arn:aws:s3:::" + s.bucketName,
					"arn:aws:s3:::" + s.bucketName + "/*",
				},
			},
		},
	}
	b, err := json.Marshal(&policy)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CreateS3Bucket(region, bucket string) (err error) {
	sess, err := session.NewSession(&aws.Config{
		Region:                        aws.String(region),
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("Failed to connect to S3: %w", err)
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
		return fmt.Errorf("Failed to connect to S3: %w", err)
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

func (s *S3Storage) listRecursive(results chan<- ListResult, dir string, filter func(string) bool) {
	prefix := filepath.Join(s.root, dir)
	// prefixes must end with / and must not end with /
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	prefix = strings.TrimPrefix(prefix, "/")

	err := s.svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket:  aws.String(s.bucketName),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int64(1000),
	}, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, value := range page.Contents {
			key := *value.Key
			if s.root != "" {
				key = strings.TrimPrefix(strings.TrimPrefix(key, s.root), "/")
			}
			if filter(key) {
				// If S3 gives us an empty/bad etag, then make it blank and cause sync instead of throwing error
				// Also, the etag includes quotes for some reason
				md5, _ := hex.DecodeString(strings.Replace(*value.ETag, "\"", "", -1))
				results <- ListResult{Path: key, MD5: md5}
			}
		}
		return true
	})
	if err != nil {
		results <- ListResult{Error: fmt.Errorf("Failed to list objects in s3://%s: %s", s.bucketName, err)}
	}
	close(results)
}

func discoverBucketRegion(bucket string) (string, error) {
	sess := session.Must(session.NewSession(&aws.Config{}))
	ctx := context.Background()
	return s3manager.GetBucketRegion(ctx, sess, bucket, "us-east-1")
}

func getBucketRegionOrCreateBucket(bucket string) (string, error) {
	// TODO (bfirsh): cache this
	region, err := discoverBucketRegion(bucket)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			// The real check for this is `aerr.Code() == s3.ErrCodeNoSuchBucket` but GetBucketRegion doesnt return right error
			if strings.Contains(aerr.Error(), "NotFound") {
				// TODO (bfirsh): report to use that this is being created, in a way that is compatible with shared library
				region = "us-east-1"
				if err := CreateS3Bucket(region, bucket); err != nil {
					return "", fmt.Errorf("Error creating bucket: %w", err)
				}
				return region, nil
			}
		}
		return "", fmt.Errorf("Failed to discover AWS region for bucket %s: %s", bucket, err)
	}
	return region, nil
}
