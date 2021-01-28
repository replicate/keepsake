package repository

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/replicate/keepsake/go/pkg/concurrency"
	"github.com/replicate/keepsake/go/pkg/console"
	"github.com/replicate/keepsake/go/pkg/errors"
	"github.com/replicate/keepsake/go/pkg/files"
)

type GCSRepository struct {
	projectID  string
	bucketName string
	root       string
	client     *storage.Client
}

func NewGCSRepository(bucket, root string) (*GCSRepository, error) {
	applicationCredentialsJSON := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON")
	options := []option.ClientOption{}
	if applicationCredentialsJSON != "" {
		jwtConfig, err := google.JWTConfigFromJSON([]byte(applicationCredentialsJSON), storage.ScopeReadWrite)
		if err != nil {
			return nil, errors.RepositoryConfigurationError(err.Error())
		}
		options = append(options, option.WithTokenSource(jwtConfig.TokenSource(context.TODO())))
	}
	client, err := storage.NewClient(context.TODO(), options...)
	if err != nil {
		return nil, errors.RepositoryConfigurationError(fmt.Sprintf("Failed to connect to Google Cloud Storage: %v", err))
	}

	return &GCSRepository{
		bucketName: bucket,
		root:       root,
		client:     client,
	}, nil
}

func (s *GCSRepository) RootURL() string {
	ret := "gs://" + s.bucketName
	if s.root != "" {
		ret += "/" + s.root
	}
	return ret
}

func (s *GCSRepository) Get(path string) ([]byte, error) {
	key := filepath.Join(s.root, path)
	pathString := fmt.Sprintf("gs://%s/%s", s.bucketName, key)
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(key)
	reader, err := obj.NewReader(context.TODO())
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, errors.DoesNotExist(fmt.Sprintf("Get: path does not exist: %s", pathString))
		}
		return nil, errors.ReadError(fmt.Sprintf("Failed to open %s: %s", pathString, err))
	}
	// FIXME: unhandled error
	defer reader.Close()
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.ReadError(fmt.Sprintf("Failed to read %s: %s", pathString, err))
	}

	return data, nil
}

// Delete deletes path. If path is a directory, it recursively deletes
// all everything under path
func (s *GCSRepository) Delete(path string) error {
	console.Debug("Deleting %s/%s...", s.RootURL(), path)
	prefix := filepath.Join(s.root, path)
	err := s.applyRecursive(prefix, func(obj *storage.ObjectHandle) error {
		return obj.Delete(context.TODO())
	})
	if err != nil {
		return errors.WriteError(fmt.Sprintf("Failed to delete %s/%s: %v", s.RootURL(), path, err))
	}
	return nil
}

// Put data at path
func (s *GCSRepository) Put(path string, data []byte) error {
	key := filepath.Join(s.root, path)
	pathString := fmt.Sprintf("gs://%s/%s", s.bucketName, key)
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(key)
	writer := obj.NewWriter(context.TODO())
	_, err := writer.Write(data)
	if err != nil {
		return errors.WriteError(fmt.Sprintf("Failed to write %q: %v", pathString, err))
	}
	if err := writer.Close(); err != nil {
		if strings.Contains(err.Error(), "notFound") {
			if err := s.ensureBucketExists(); err != nil {
				return err
			}
			writer := obj.NewWriter(context.TODO())
			_, err := writer.Write(data)
			if err != nil {
				return errors.WriteError(fmt.Sprintf("Failed to write %q: %v", pathString, err))
			}
			if err := writer.Close(); err != nil {
				return errors.WriteError(fmt.Sprintf("Failed to write %q: %v", pathString, err))
			}
			return nil
		}
		return errors.WriteError(fmt.Sprintf("Failed to write %q: %v", pathString, err))
	}
	return nil
}

func (s *GCSRepository) PutPath(localPath string, repoPath string) error {
	files, err := getListOfFilesToPut(localPath, filepath.Join(s.root, repoPath))
	if err != nil {
		return err
	}
	bucket := s.client.Bucket(s.bucketName)
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
			return errors.WriteError(err.Error())
		}
	}
	if err := queue.Wait(); err != nil {
		return errors.WriteError(err.Error())
	}
	return nil
}

func (s *GCSRepository) PutPathTar(localPath, tarPath, includePath string) error {
	if !strings.HasSuffix(tarPath, ".tar.gz") {
		return fmt.Errorf("PutPathTar: tarPath must end with .tar.gz")
	}
	if err := s.ensureBucketExists(); err != nil {
		return err
	}

	key := filepath.Join(s.root, tarPath)
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(key)
	writer := obj.NewWriter(context.TODO())

	if err := putPathTar(localPath, writer, filepath.Base(tarPath), includePath); err != nil {
		return errors.WriteError(err.Error())
	}
	if err := writer.Close(); err != nil {
		return errors.WriteError(err.Error())
	}
	return nil
}

// List files in a path non-recursively
func (s *GCSRepository) List(dir string) ([]string, error) {
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
			return nil, errors.ReadError(fmt.Sprintf("Failed to list %s/%s: %s", s.RootURL(), dir, err))
		}
		p := attrs.Name
		if s.root != "" {
			p = strings.TrimPrefix(strings.TrimPrefix(p, s.root), "/")
		}
		if p != "" {
			results = append(results, p)
		}
	}
	return results, nil
}

func (s *GCSRepository) ListTarFile(tarPath string) ([]string, error) {
	// archiver doesn't let us use readers, so download to temporary file
	// TODO: make a better tar implementation
	tmpdir, err := files.TempDir("tar")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpdir)
	tmptarball := filepath.Join(tmpdir, filepath.Base(tarPath))
	if err := s.GetPath(tarPath, tmptarball); err != nil {
		return nil, err
	}
	exists, err := files.FileExists(tmptarball)
	if err != nil {
		return []string{}, err
	}
	if !exists {
		return nil, errors.DoesNotExist("Path does not exist: " + tmptarball)
	}

	files, err := getListOfFilesInTar(tmptarball)
	if err != nil {
		return []string{}, err
	}

	tarname := filepath.Base(strings.TrimSuffix(tarPath, ".tar.gz"))
	for idx := range files {
		files[idx] = strings.TrimPrefix(files[idx], tarname+"/")
	}

	return files, nil
}

// List files in a path recursively
func (s *GCSRepository) ListRecursive(results chan<- ListResult, dir string) {
	s.listRecursive(results, dir, func(_ string) bool { return true })
}

func (s *GCSRepository) MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string) {
	s.listRecursive(results, folder, func(key string) bool {
		return filepath.Base(key) == filename
	})
}

func (s *GCSRepository) listRecursive(results chan<- ListResult, dir string, filter func(string) bool) {
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
			break
		}
		if err != nil {
			// Treat non-existent buckets as empty
			// Can't figure out how to check this error more strongly
			if strings.Contains(err.Error(), "storage: bucket doesn't exist") {
				break
			}

			results <- ListResult{Error: fmt.Errorf("Failed to list gs://%s/%s: %s", s.bucketName, prefix, err)}
			break
		}
		if filter(attrs.Name) {
			p := attrs.Name
			if s.root != "" {
				p = strings.TrimPrefix(strings.TrimPrefix(p, s.root), "/")
			}
			results <- ListResult{Path: p, MD5: attrs.MD5}
		}
	}
	close(results)
}

// GetPath recursively copies repoDir to localDir
func (s *GCSRepository) GetPath(repoDir string, localDir string) error {
	prefix := filepath.Join(s.root, repoDir)
	err := s.applyRecursive(prefix, func(obj *storage.ObjectHandle) error {
		gcsPathString := fmt.Sprintf("gs://%s/%s", s.bucketName, obj.ObjectName())
		reader, err := obj.NewReader(context.TODO())
		if err != nil {
			return errors.ReadError(fmt.Sprintf("Failed to open %s: %v", gcsPathString, err))
		}
		defer reader.Close()

		relPath, err := filepath.Rel(prefix, obj.ObjectName())
		if err != nil {
			return errors.ReadError(fmt.Sprintf("Failed to determine directory of %s relative to %s: %v", obj.ObjectName(), repoDir, err))
		}
		localPath := filepath.Join(localDir, relPath)
		localDir := filepath.Dir(localPath)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return errors.ReadError(fmt.Sprintf("Failed to create directory %s: %v", localDir, err))
		}

		f, err := os.Create(localPath)
		if err != nil {
			return errors.ReadError(fmt.Sprintf("Failed to create file %s: %v", localPath, err))
		}
		defer f.Close()

		console.Debug("Downloading %s to %s", gcsPathString, localPath)
		if _, err := io.Copy(f, reader); err != nil {
			return errors.ReadError(fmt.Sprintf("Failed to copy %s to %s: %v", gcsPathString, localPath, err))
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Failed to copy gs://%s/%s to %s: %v", s.bucketName, repoDir, localDir, err)
	}
	return nil
}

func (s *GCSRepository) GetPathTar(tarPath, localPath string) error {
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
	exists, err := files.FileExists(tmptarball)
	if err != nil {
		return err
	}
	if !exists {
		return errors.DoesNotExist(fmt.Sprintf("Path does not exist: %s", tmptarball))
	}
	return extractTar(tmptarball, localPath)
}

func (s *GCSRepository) GetPathItemTar(tarPath, itemPath, localPath string) error {
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
	exists, err := files.FileExists(tmptarball)
	if err != nil {
		return err
	}
	if !exists {
		return errors.DoesNotExist("Path does not exist: " + tmptarball)
	}
	return extractTarItem(tmptarball, itemPath, localPath)
}

func (s *GCSRepository) bucketExists() (bool, error) {
	bucket := s.client.Bucket(s.bucketName)
	_, err := bucket.Attrs(context.TODO())
	if err == nil {
		return true, nil
	}
	if err == storage.ErrBucketNotExist {
		return false, nil
	}
	return false, errors.RepositoryConfigurationError(fmt.Sprintf("Failed to determine if bucket gs://%s exists: %v", s.bucketName, err))
}

func (s *GCSRepository) ensureBucketExists() error {
	exists, err := s.bucketExists()
	if err != nil {
		return err
	}
	if !exists {
		return s.CreateBucket()
	}
	return nil
}

func (s *GCSRepository) CreateBucket() error {
	projectID, err := s.getProjectID()
	if err != nil {
		return err
	}
	bucket := s.client.Bucket(s.bucketName)
	if err := bucket.Create(context.TODO(), projectID, nil); err != nil {
		return fmt.Errorf("Failed to create bucket gs://%s: %v", s.bucketName, err)
	}
	return nil
}

// Note: prefix does not include s.root
func (s *GCSRepository) applyRecursive(prefix string, fn func(obj *storage.ObjectHandle) error) error {
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
func (s *GCSRepository) getProjectID() (string, error) {
	if os.Getenv("GOOGLE_CLOUD_PROJECT") != "" {
		return os.Getenv("GOOGLE_CLOUD_PROJECT"), nil
	}
	if s.projectID == "" {
		projectID, err := discoverProjectID()
		if err != nil {
			return "", err
		}
		s.projectID = projectID
	}
	return s.projectID, nil
}

func discoverProjectID() (string, error) {
	cmd := exec.Command("gcloud", "config", "config-helper", "--format=value(configuration.properties.core.project)")
	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		if ee, ok := err.(*exec.ExitError); ok {
			stderr += "\n" + string(ee.Stderr)
		}
		return "", errors.RepositoryConfigurationError(fmt.Sprintf("Failed to determine default GCP project (using gcloud config config-helper): %v\n%s", err, stderr))
	}
	return strings.TrimSpace(string(out)), nil
}
