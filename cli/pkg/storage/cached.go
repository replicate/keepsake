package storage

import (
	"path"
	"strings"

	"replicate.ai/cli/pkg/console"
)

// CachedStorage wraps another storage, caching a prefix in a local directory.
//
// SyncCache() syncs cachePrefix locally, which you must call before doing any
// reads. It is not done automatically so you can control output to the user about
// syncing.
//
// If a read hits a path starting with cachePrefix, it will use the local cached version.
type CachedStorage struct {
	storage      Storage
	cachePrefix  string
	cacheDir     string
	cacheStorage *DiskStorage
	isSynced     bool
}

func NewCachedStorage(store Storage, cachePrefix string, cacheDir string) (*CachedStorage, error) {
	// This doesn't actually return an error, but catch in case of future errors
	cacheStorage, err := NewDiskStorage(cacheDir)
	if err != nil {
		return nil, err
	}
	return &CachedStorage{
		storage:      store,
		cachePrefix:  cachePrefix,
		cacheDir:     cacheDir,
		cacheStorage: cacheStorage,
		isSynced:     false,
	}, nil
}

// NewCachedMetadataStorage returns a CachedStorage that caches the metadata/ path in
// .replicate/metadata-cache in a source dir
func NewCachedMetadataStorage(store Storage, projectDir string) (*CachedStorage, error) {
	return NewCachedStorage(store, "metadata", path.Join(projectDir, ".replicate/metadata-cache"))
}

func (s *CachedStorage) Get(p string) ([]byte, error) {
	if strings.HasPrefix(p, s.cachePrefix) {
		return s.cacheStorage.Get(p)
	}
	return s.storage.Get(p)
}

func (s *CachedStorage) Put(p string, data []byte) error {
	// FIXME: potential for cache and remote to get out of sync on error
	if strings.HasPrefix(p, s.cachePrefix) {
		if err := s.cacheStorage.Put(p, data); err != nil {
			return err
		}
	}
	return s.storage.Put(p, data)
}

func (s *CachedStorage) GetPath(storagePath string, localPath string) error {
	if strings.HasPrefix(storagePath, s.cachePrefix) {
		return s.cacheStorage.GetPath(storagePath, localPath)
	}
	return s.storage.GetPath(storagePath, localPath)
}

func (s *CachedStorage) GetPathTar(tarPath, localPath string) error {
	if strings.HasPrefix(tarPath, s.cachePrefix) {
		return s.cacheStorage.GetPathTar(tarPath, localPath)
	}
	return s.storage.GetPathTar(tarPath, localPath)
}

func (s *CachedStorage) PutPath(localPath string, storagePath string) error {
	// FIXME: potential for cache and remote to get out of sync on error
	if strings.HasPrefix(storagePath, s.cachePrefix) {
		if err := s.cacheStorage.PutPath(localPath, storagePath); err != nil {
			return err
		}
	}
	return s.storage.PutPath(localPath, storagePath)

}

func (s *CachedStorage) PutPathTar(localPath, tarPath, includePath string) error {
	// FIXME: potential for cache and remote to get out of sync on error
	if strings.HasPrefix(tarPath, s.cachePrefix) {
		if err := s.cacheStorage.PutPathTar(localPath, tarPath, includePath); err != nil {
			return err
		}
	}
	return s.storage.PutPathTar(localPath, tarPath, includePath)
}

func (s *CachedStorage) List(p string) ([]string, error) {
	if strings.HasPrefix(p, s.cachePrefix) {
		return s.cacheStorage.List(p)
	}
	return s.storage.List(p)
}

func (s *CachedStorage) ListRecursive(results chan<- ListResult, path string) {
	if strings.HasPrefix(path, s.cachePrefix) {
		s.cacheStorage.ListRecursive(results, path)
		return
	}
	s.storage.ListRecursive(results, path)
}

func (s *CachedStorage) MatchFilenamesRecursive(results chan<- ListResult, path string, filename string) {
	if strings.HasPrefix(path, s.cachePrefix) {
		s.cacheStorage.MatchFilenamesRecursive(results, path, filename)
		return
	}
	s.storage.MatchFilenamesRecursive(results, path, filename)
}

func (s *CachedStorage) Delete(p string) error {
	if strings.HasPrefix(p, s.cachePrefix) {
		if err := s.cacheStorage.Delete(p); err != nil {
			return err
		}
	}
	return s.storage.Delete(p)
}

func (s *CachedStorage) RootURL() string {
	return s.storage.RootURL()
}

func (s *CachedStorage) SyncCache() error {
	console.Debug("Syncing %s/%s to %s/%s", s.storage.RootURL(), s.cachePrefix, s.cacheStorage.RootURL(), s.cachePrefix)
	return Sync(s.storage, s.cachePrefix, s.cacheStorage, s.cachePrefix)
}
