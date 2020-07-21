package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type GCSStorage struct {
	bucketName string
	client     *storage.Client
}

func NewGCSStorage(bucket string) (*GCSStorage, error) {
	client, err := storage.NewClient(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Google Cloud Storage, got error: %s", err)
	}

	return &GCSStorage{
		bucketName: bucket,
		client:     client,
	}, nil
}

func (s *GCSStorage) Get(path string) ([]byte, error) {
	pathString := fmt.Sprintf("gs://%s/%s", s.bucketName, path)
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(path)
	reader, err := obj.NewReader(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("Failed to open %s, got error: %s", pathString, err)
	}
	defer reader.Close()
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s, got error: %s", pathString, err)
	}

	return data, nil
}

// Put data at path
func (s *GCSStorage) Put(path string, data []byte) error {
	// TODO
	return nil
}

func (s *GCSStorage) MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string) {
	s.listRecursive(results, folder, func(key string) bool {
		return filepath.Base(key) == filename
	})
}

func (s *GCSStorage) listRecursive(results chan<- ListResult, folder string, filter func(string) bool) {
	// prefixes must end with / and must not end with /
	if !strings.HasSuffix(folder, "/") {
		folder += "/"
	}
	folder = strings.TrimPrefix(folder, "/")

	bucket := s.client.Bucket(s.bucketName)
	it := bucket.Objects(context.TODO(), &storage.Query{
		Prefix: folder,
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			close(results)
			break
		}
		if err != nil {
			results <- ListResult{Error: fmt.Errorf("Failed to list gs://%s/%s", s.bucketName, folder)}
		}
		if filter(attrs.Name) {
			results <- ListResult{Path: attrs.Name}
		}
	}
}
