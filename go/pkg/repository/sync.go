package repository

import (
	"bytes"
	"context"
	"path"
	"strings"

	"github.com/replicate/keepsake/go/pkg/concurrency"
)

// Sync destRepository/destPath to match sourceRepository/sourcePath
//
// - If file exists in source, but not in dest, it will copy from source to dest
// - If file exists in both but different content, it will copy from source to dest
// - If file exists in dest but not in source, it will delete in dest
func Sync(sourceRepository Repository, sourcePath string, destRepository Repository, destPath string) error {
	// A queue to use for the various storage operations we have to run
	queue := concurrency.NewWorkerQueue(context.Background(), maxWorkers)

	// 1: Fetch destFiles synchronously off disk
	// TODO: This could be optimized by doing this while source list request is in flight
	results := make(chan ListResult)
	// path -> MD5 hash map used to efficiently check if files should be synced
	destFiles := make(map[string][]byte)

	go destRepository.ListRecursive(results, destPath)
	for result := range results {
		if result.Error != nil {
			return result.Error
		}
		destFiles[strings.TrimPrefix(result.Path, destPath)] = result.MD5
	}

	// 2: Copy files from dest to source which don't exist or have changed
	sourceFiles := make(chan ListResult)
	// path map used for step (3)
	sourceFileMap := make(map[string]struct{})
	go sourceRepository.ListRecursive(sourceFiles, sourcePath)
	for sourceFile := range sourceFiles {
		if sourceFile.Error != nil {
			return sourceFile.Error
		}
		relativePath := strings.TrimPrefix(sourceFile.Path, sourcePath)
		sourceFileMap[relativePath] = struct{}{}

		needsCopying := false
		if destMD5, found := destFiles[relativePath]; found {
			if !bytes.Equal(sourceFile.MD5, destMD5) {
				needsCopying = true
			}
		} else {
			needsCopying = true
		}

		if needsCopying {
			// Variables used in closure
			sourcePath := sourcePath
			destPath := destPath
			relativePath := relativePath
			err := queue.Go(func() error {
				data, err := sourceRepository.Get(path.Join(sourcePath, relativePath))
				if err != nil {
					return err
				}
				return destRepository.Put(path.Join(destPath, relativePath), data)
			})
			if err != nil {
				return err
			}
		}
	}

	// 3: Delete files from dest that don't exist in source
	for relativePath := range destFiles {
		if _, found := sourceFileMap[relativePath]; !found {
			// Variables used in closure
			destPath := destPath
			relativePath := relativePath
			err := queue.Go(func() error {
				return destRepository.Delete(path.Join(destPath, relativePath))
			})
			if err != nil {
				return err
			}
		}
	}

	return queue.Wait()
}
