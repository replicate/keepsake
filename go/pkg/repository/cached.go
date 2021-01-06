package repository

import (
	"strings"

	"github.com/replicate/replicate/go/pkg/console"
)

// CachedRepository wraps another repository, caching a prefix in a local directory.
//
// SyncCache() syncs cachePrefix locally, which you must call before doing any
// reads. It is not done automatically so you can control output to the user about
// syncing.
//
// If a read hits a path starting with cachePrefix, it will use the local cached version.
type CachedRepository struct {
	repository      Repository
	cachePrefix     string
	cacheDir        string
	cacheRepository *DiskRepository
	isSynced        bool
}

func NewCachedRepository(repo Repository, cachePrefix string, projectDir string, cacheDir string) (*CachedRepository, error) {
	// This doesn't actually return an error, but catch in case of future errors
	cacheRepository, err := NewDiskRepository(cacheDir)
	if err != nil {
		return nil, err
	}
	return &CachedRepository{
		repository:      repo,
		cachePrefix:     cachePrefix,
		cacheDir:        cacheDir,
		cacheRepository: cacheRepository,
		isSynced:        false,
	}, nil
}

// NewCachedMetadataRepository returns a CachedRepository that caches the metadata/ path in
// .replicate/metadata-cache in a source dir
func NewCachedMetadataRepository(projectDir string, repo Repository) (*CachedRepository, error) {
	return NewCachedRepository(repo, "metadata", projectDir, ".replicate/metadata-cache")
}

func (s *CachedRepository) Get(p string) ([]byte, error) {
	if strings.HasPrefix(p, s.cachePrefix) {
		return s.cacheRepository.Get(p)
	}
	return s.repository.Get(p)
}

func (s *CachedRepository) Put(p string, data []byte) error {
	// FIXME: potential for cache and remote to get out of sync on error
	if strings.HasPrefix(p, s.cachePrefix) {
		if err := s.cacheRepository.Put(p, data); err != nil {
			return err
		}
	}
	return s.repository.Put(p, data)
}

func (s *CachedRepository) GetPath(repoPath string, localPath string) error {
	if strings.HasPrefix(repoPath, s.cachePrefix) {
		return s.cacheRepository.GetPath(repoPath, localPath)
	}
	return s.repository.GetPath(repoPath, localPath)
}

func (s *CachedRepository) GetPathTar(tarPath, localPath string) error {
	if strings.HasPrefix(tarPath, s.cachePrefix) {
		return s.cacheRepository.GetPathTar(tarPath, localPath)
	}
	return s.repository.GetPathTar(tarPath, localPath)
}

func (s *CachedRepository) GetPathItemTar(tarPath, itemPath, localPath string) error {
	if strings.HasPrefix(tarPath, s.cachePrefix) {
		return s.cacheRepository.GetPathTar(tarPath, localPath)
	}
	return s.repository.GetPathItemTar(tarPath, itemPath, localPath)
}

func (s *CachedRepository) PutPath(localPath string, repoPath string) error {
	// FIXME: potential for cache and remote to get out of sync on error
	if strings.HasPrefix(repoPath, s.cachePrefix) {
		if err := s.cacheRepository.PutPath(localPath, repoPath); err != nil {
			return err
		}
	}
	return s.repository.PutPath(localPath, repoPath)

}

func (s *CachedRepository) PutPathTar(localPath, tarPath, includePath string) error {
	// FIXME: potential for cache and remote to get out of sync on error
	if strings.HasPrefix(tarPath, s.cachePrefix) {
		if err := s.cacheRepository.PutPathTar(localPath, tarPath, includePath); err != nil {
			return err
		}
	}
	return s.repository.PutPathTar(localPath, tarPath, includePath)
}

func (s *CachedRepository) List(p string) ([]string, error) {
	if strings.HasPrefix(p, s.cachePrefix) {
		return s.cacheRepository.List(p)
	}
	return s.repository.List(p)
}

func (s *CachedRepository) ListTarFile(p string) ([]string, error) {
	if strings.HasPrefix(p, s.cachePrefix) {
		return s.cacheRepository.List(p)
	}
	return s.repository.ListTarFile(p)
}

func (s *CachedRepository) ListRecursive(results chan<- ListResult, path string) {
	if strings.HasPrefix(path, s.cachePrefix) {
		s.cacheRepository.ListRecursive(results, path)
		return
	}
	s.repository.ListRecursive(results, path)
}

func (s *CachedRepository) MatchFilenamesRecursive(results chan<- ListResult, path string, filename string) {
	if strings.HasPrefix(path, s.cachePrefix) {
		s.cacheRepository.MatchFilenamesRecursive(results, path, filename)
		return
	}
	s.repository.MatchFilenamesRecursive(results, path, filename)
}

func (s *CachedRepository) Delete(p string) error {
	if strings.HasPrefix(p, s.cachePrefix) {
		if err := s.cacheRepository.Delete(p); err != nil {
			return err
		}
	}
	return s.repository.Delete(p)
}

func (s *CachedRepository) RootURL() string {
	return s.repository.RootURL()
}

func (s *CachedRepository) SyncCache() error {
	console.Debug("Syncing %s/%s to %s/%s", s.repository.RootURL(), s.cachePrefix, s.cacheRepository.RootURL(), s.cachePrefix)
	return Sync(s.repository, s.cachePrefix, s.cacheRepository, s.cachePrefix)
}
