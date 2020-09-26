package storage

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/adjust/uniuri"
	iam "google.golang.org/api/iam/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/serviceusage/v1"

	"replicate.ai/cli/pkg/cache"
	"replicate.ai/cli/pkg/concurrency"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/slices"
)

const waitRefreshInterval = 500 * time.Millisecond

var requiredServices = []string{
	"storage-api.googleapis.com",
	"iam.googleapis.com",
}

type GCSStorage struct {
	projectID  string
	bucketName string
	root       string
	client     *storage.Client
}

func NewGCSStorage(bucket, root string) (*GCSStorage, error) {
	options := []option.ClientOption{}

	// When inside `replicate run`, get the options passed from PrepareRunEnv
	key := os.Getenv("REPLICATE_GCP_SERVICE_ACCOUNT_KEY")
	if key != "" {
		keyJSON, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			return nil, fmt.Errorf("Error decoding REPLICATE_GCP_SERVICE_ACCOUNT_KEY: %w", err)
		}
		options = append(options, option.WithCredentialsJSON(keyJSON))
	}

	client, err := storage.NewClient(context.TODO(), options...)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Google Cloud Storage: %w", err)
	}

	return &GCSStorage{
		bucketName: bucket,
		root:       root,
		client:     client,
		// When inside `replicate run`, default to project passed from PrepareRunEnv
		projectID: os.Getenv("REPLICATE_GCP_PROJECT"),
	}, nil
}

func (s *GCSStorage) RootURL() string {
	ret := "gs://" + s.bucketName
	if s.root != "" {
		ret += "/" + s.root
	}
	return ret
}

func (s *GCSStorage) RootExists() (bool, error) {
	bucket := s.client.Bucket(s.bucketName)
	_, err := bucket.Attrs(context.TODO())
	if err == nil {
		return true, nil
	}
	if err == storage.ErrBucketNotExist {
		return false, nil
	}
	return false, fmt.Errorf("Failed to determine if bucket gs://%s exists: %w", s.bucketName, err)
}

func (s *GCSStorage) Get(path string) ([]byte, error) {
	key := filepath.Join(s.root, path)
	pathString := fmt.Sprintf("gs://%s/%s", s.bucketName, key)
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(key)
	reader, err := obj.NewReader(context.TODO())
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, &DoesNotExistError{msg: "Get: path does not exist: " + pathString}
		}
		return nil, fmt.Errorf("Failed to open %s, got error: %s", pathString, err)
	}
	// FIXME: unhandled error
	defer reader.Close()
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s, got error: %s", pathString, err)
	}

	return data, nil
}

// Delete deletes path. If path is a directory, it recursively deletes
// all everything under path
func (s *GCSStorage) Delete(path string) error {
	console.Debug("Deleting %s/%s...", s.RootURL(), path)
	prefix := filepath.Join(s.root, path)
	err := s.applyRecursive(prefix, func(obj *storage.ObjectHandle) error {
		return obj.Delete(context.TODO())
	})
	if err != nil {
		return fmt.Errorf("Failed to delete %s/%s: %w", s.RootURL(), path, err)
	}
	return nil
}

// Put data at path
func (s *GCSStorage) Put(path string, data []byte) error {
	key := filepath.Join(s.root, path)
	pathString := fmt.Sprintf("gs://%s/%s", s.bucketName, key)
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(key)
	writer := obj.NewWriter(context.TODO())
	_, err := writer.Write(data)
	if err != nil {
		return fmt.Errorf("Failed to write %q, got error: %w", pathString, err)
	}
	if err := writer.Close(); err != nil {
		if strings.Contains(err.Error(), "notFound") {
			if err := s.EnsureBucketExists(); err != nil {
				return fmt.Errorf("Error creating bucket: %w", err)
			}
			writer := obj.NewWriter(context.TODO())
			_, err := writer.Write(data)
			if err != nil {
				return fmt.Errorf("Failed to write %q, got error: %w", pathString, err)
			}
			if err := writer.Close(); err != nil {
				return fmt.Errorf("Failed to write %q, got error: %w", pathString, err)
			}
			return nil
		}
		return fmt.Errorf("Failed to write %q, got error: %w", pathString, err)
	}
	return nil
}

func (s *GCSStorage) PutDirectory(localPath string, storagePath string) error {
	files, err := putDirectoryFiles(localPath, filepath.Join(s.root, storagePath))
	bucket := s.client.Bucket(s.bucketName)
	if err != nil {
		return err
	}
	queue := concurrency.NewWorkerQueue(context.Background(), maxWorkers)
	for _, file := range files {
		// Variables used in closure
		file := file
		err := queue.Go(func() error {
			writer := bucket.Object(file.Dest).NewWriter(context.TODO())

			reader, err := os.Open(file.Source)
			if err != nil {
				return err
			}
			if _, err := io.Copy(writer, reader); err != nil {
				return err
			}
			if err := reader.Close(); err != nil {
				return err
			}
			if err := writer.Close(); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return queue.Wait()
}

// List files in a path non-recursively
func (s *GCSStorage) List(dir string) ([]string, error) {
	results := []string{}
	prefix := filepath.Join(s.root, dir)

	// prefixes must end with / and must not end with /
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	prefix = strings.TrimPrefix(prefix, "/")

	bucket := s.client.Bucket(s.bucketName)
	it := bucket.Objects(context.TODO(), &storage.Query{
		Prefix:    prefix,
		Delimiter: "/",
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Failed to list %s/%s", s.RootURL(), dir)
		}
		p := attrs.Name
		if s.root != "" {
			p = strings.TrimPrefix(strings.TrimPrefix(p, s.root), "/")
		}
		results = append(results, p)
	}
	return results, nil
}

// List files in a path recursively
func (s *GCSStorage) ListRecursive(results chan<- ListResult, dir string) {
	s.listRecursive(results, dir, func(_ string) bool { return true })
}

func (s *GCSStorage) MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string) {
	s.listRecursive(results, folder, func(key string) bool {
		return filepath.Base(key) == filename
	})
}

func (s *GCSStorage) listRecursive(results chan<- ListResult, dir string, filter func(string) bool) {
	prefix := filepath.Join(s.root, dir)
	// prefixes must end with / and must not end with /
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	prefix = strings.TrimPrefix(prefix, "/")

	bucket := s.client.Bucket(s.bucketName)
	it := bucket.Objects(context.TODO(), &storage.Query{
		Prefix: prefix,
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			close(results)
			break
		}
		if err != nil {
			results <- ListResult{Error: fmt.Errorf("Failed to list gs://%s/%s", s.bucketName, prefix)}
		}
		if filter(attrs.Name) {
			p := attrs.Name
			if s.root != "" {
				p = strings.TrimPrefix(strings.TrimPrefix(p, s.root), "/")
			}
			results <- ListResult{Path: p}
		}
	}
}

// GetDirectory recursively copies storageDir to localDir
func (s *GCSStorage) GetDirectory(storageDir string, localDir string) error {
	prefix := filepath.Join(s.root, storageDir)
	err := s.applyRecursive(prefix, func(obj *storage.ObjectHandle) error {
		gcsPathString := fmt.Sprintf("gs://%s/%s", s.bucketName, obj.ObjectName())
		reader, err := obj.NewReader(context.TODO())
		if err != nil {
			return fmt.Errorf("Failed to open %s, got error: %w", gcsPathString, err)
		}
		defer reader.Close()

		relPath, err := filepath.Rel(prefix, obj.ObjectName())
		if err != nil {
			return fmt.Errorf("Failed to determine directory of %s relative to %s, got error: %w", obj.ObjectName(), storageDir, err)
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
		defer f.Close()

		console.Debug("Downloading %s to %s", gcsPathString, localPath)
		if _, err := io.Copy(f, reader); err != nil {
			return fmt.Errorf("Failed to copy %s to %s, got error: %w", gcsPathString, localPath, err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Failed to copy gs://%s/%s to %s, got error: %w", s.bucketName, storageDir, localDir, err)
	}
	return nil
}

func (s *GCSStorage) EnsureBucketExists() error {
	exists, err := s.RootExists()
	if err != nil {
		return err
	}
	if !exists {
		return s.CreateBucket()
	}
	return nil
}

func (s *GCSStorage) CreateBucket() error {
	projectID, err := s.getProjectID()
	if err != nil {
		return err
	}
	bucket := s.client.Bucket(s.bucketName)
	if err := bucket.Create(context.TODO(), projectID, nil); err != nil {
		return fmt.Errorf("Failed to create bucket gs://%s: %w", s.bucketName, err)
	}
	return nil
}

// PrepareRunEnv gets or creates the GCS bucket and service account,
// and returns ["REPLICATE_GCP_SERVICE_ACCOUNT_KEY=<key>",
// "REPLICATE_GCP_PROJECT=<project>"], where 'key' is a base64-encoded
// json key
//
// This is used by `replicate run` to pass credentials to Replicate
// running inside the container.
func (s *GCSStorage) PrepareRunEnv() ([]string, error) {
	if err := s.EnsureBucketExists(); err != nil {
		return nil, err
	}
	serviceAccountKey, err := s.GetOrCreateServiceAccount()
	if err != nil {
		return nil, err
	}
	projectID, err := s.getProjectID()
	if err != nil {
		return nil, err
	}
	return []string{"REPLICATE_GCP_SERVICE_ACCOUNT_KEY=" + serviceAccountKey, "REPLICATE_GCP_PROJECT=" + projectID}, nil
}

// GetOrCreateServiceAccount returns a base64-encoded service account
// json key. If a cached version exists it is returned, otherwise a
// new service account is created and the key is cached
// TODO(andreas): since this is sensitive data, it should perhaps be cached in a more secure way (e.g. keychain or google cloud secrets manager)
func (s *GCSStorage) GetOrCreateServiceAccount() (serviceAccountKey string, err error) {
	cacheKey := "service-account-" + s.bucketName
	if key, ok := cache.GetString(cacheKey); ok {
		return key, nil
	}
	key, err := s.createServiceAccount()
	if err != nil {
		return "", err
	}
	if err := cache.SetString(cacheKey, key); err != nil {
		return "", err
	}
	return key, nil
}

// CreateServiceAccount creates a new service account, gives it OWNER
// rights to the bucket, and returns the base64-encoded service
// account json key
func (s *GCSStorage) createServiceAccount() (serviceAccountKey string, err error) {
	if err := s.enableRequiredServices(); err != nil {
		// only warn and move on. it's possible that the user
		// has permissions to create service accounts, without
		// permission to list or enable services.
		console.Warn("Failed to enable Google Cloud services: %s\n\nIf the following Google Cloud services are not enabled, Google Cloud Storage may not work: %s", err, strings.Join(requiredServices, ", "))
	}

	projectID, err := s.getProjectID()
	if err != nil {
		return "", err
	}

	maxBucketNameLength := 30 - len("replicate") - 7 - 2
	bucketName := s.bucketName
	if len(bucketName) > maxBucketNameLength {
		bucketName = bucketName[:maxBucketNameLength]
	}
	name := fmt.Sprintf("replicate-%s-%s", bucketName, strings.ToLower(uniuri.NewLen(7)))

	iamService, err := iam.NewService(context.Background())
	if err != nil {
		return "", fmt.Errorf("Failed to connect to Google Cloud IAM: %w", err)
	}

	console.Debug("Creating service account %s as an owner of gs://%s", name, s.bucketName)
	accountRequest := &iam.CreateServiceAccountRequest{
		AccountId: name,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: "replicate/" + s.bucketName,
		},
	}
	account, err := iamService.Projects.ServiceAccounts.Create("projects/"+projectID, accountRequest).Do()
	if err != nil {
		return "", fmt.Errorf("Failed to create Google Cloud service account: %w", err)
	}

	resource := fmt.Sprintf("projects/%s/serviceAccounts/%s", projectID, account.Email)
	keyRequest := &iam.CreateServiceAccountKeyRequest{}
	key, err := iamService.Projects.ServiceAccounts.Keys.Create(resource, keyRequest).Do()
	if err != nil {
		return "", fmt.Errorf("Failed to create Google Cloud service account key for %s: %w", account.Email, err)
	}

	bucket := s.client.Bucket(s.bucketName)
	err = bucket.ACL().Set(context.TODO(), storage.ACLEntity("user-"+account.Email), storage.RoleOwner)
	if err != nil {
		return "", fmt.Errorf("Failed to make service account %s an owner of %s: %w", account.Email, s.bucketName, err)
	}

	return key.PrivateKeyData, nil
}

// Note: prefix does not include s.root
func (s *GCSStorage) applyRecursive(prefix string, fn func(obj *storage.ObjectHandle) error) error {
	queue := concurrency.NewWorkerQueue(context.Background(), maxWorkers)

	bucket := s.client.Bucket(s.bucketName)
	it := bucket.Objects(context.TODO(), &storage.Query{
		Prefix: prefix,
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		err = queue.Go(func() error {
			obj := bucket.Object(attrs.Name)
			return fn(obj)
		})
		if err != nil {
			return err
		}
	}
	return queue.Wait()
}

// getProjectID shells out to gcloud config config-helper to get
// the project ID. This is the recommended way
// https://github.com/googleapis/google-cloud-go/issues/707
func (s *GCSStorage) getProjectID() (string, error) {
	if s.projectID == "" {
		projectID, err := discoverProjectID()
		if err != nil {
			return "", err
		}
		s.projectID = projectID
	}
	return s.projectID, nil
}

func (s *GCSStorage) enableRequiredServices() error {
	projectID, err := s.getProjectID()
	if err != nil {
		return err
	}

	serviceusageClient, err := serviceusage.NewService(context.TODO())
	if err != nil {
		return fmt.Errorf("Failed to create GCP ServiceUsage client: %w", err)
	}

	enabledServices := []string{}
	err = serviceusageClient.Services.List("projects/"+projectID).
		PageSize(200).
		Filter("state:ENABLED").
		Pages(context.TODO(), func(resp *serviceusage.ListServicesResponse) error {
			for _, service := range resp.Services {
				name := service.Name
				shortName := strings.Split(name, "/services/")[1]
				enabledServices = append(enabledServices, shortName)
			}
			return nil
		})
	if err != nil {
		return fmt.Errorf("Failed to list enabled Google Cloud APIs: %w", err)
	}

	newServices := []string{}
	for _, name := range requiredServices {
		if !slices.ContainsString(enabledServices, name) {
			newServices = append(newServices, name)
		}
	}
	if len(newServices) == 0 {
		return nil
	}

	console.Info("Enabling Google Cloud APIs: %s", strings.Join(newServices, ", "))

	op, err := serviceusageClient.Services.BatchEnable("projects/"+projectID, &serviceusage.BatchEnableServicesRequest{ServiceIds: newServices}).Do()
	if err != nil {
		return fmt.Errorf("Failed to enable required APIs, got error: %s", err)
	}
	return waitForOperation(context.TODO(), func() (bool, error) {
		op, err = serviceusageClient.Operations.Get(op.Name).Do()
		if err != nil {
			return false, err
		}
		if op.Error != nil {
			return false, fmt.Errorf("%v", op.Error)
		}
		if op.Done {
			return true, nil
		}
		return false, nil
	})
}

func discoverProjectID() (string, error) {
	cmd := exec.Command("gcloud", "config", "config-helper", "--format=value(configuration.properties.core.project)")
	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		if ee, ok := err.(*exec.ExitError); ok {
			stderr += "\n" + string(ee.Stderr)
		}
		return "", fmt.Errorf("Failed to determine default GCP project (using gcloud config config-helper): %w\n%s", err, stderr)
	}
	return strings.TrimSpace(string(out)), nil
}

func waitForOperation(ctx context.Context, isDone func() (bool, error)) error {
	ticker := time.NewTicker(waitRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for operation to complete")
		case <-ticker.C:
			done, err := isDone()
			if err != nil {
				return fmt.Errorf("Operation failed: %s", err)
			}

			if done {
				return nil
			}
		}
	}
}
