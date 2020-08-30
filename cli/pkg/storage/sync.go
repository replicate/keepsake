package storage

import (
	"context"
	"path"
	"strings"

	"replicate.ai/cli/pkg/concurrency"
	"replicate.ai/cli/pkg/console"
)

// Sync destStorage/destPath to match sourceStorage/sourcePath
//
// It does this based on filename. If file exists in source, but not in dest, it will copy it from
// source to dest. If it exists in dest but not in source, it will delete it in dest.
//
// This could be improved by:
// - checking mtime
// - make better use of the channel to list files, so it can start to copy while it's paginating
func Sync(sourceStorage Storage, sourcePath string, destStorage Storage, destPath string) error {
	rootExists, err := sourceStorage.RootExists()
	if err != nil {
		return err
	}
	if !rootExists {
		console.Debug("Not syncing non-existing storage at %s", sourceStorage.RootURL())
		return nil
	}

	sourceFiles, err := listRecursiveSlice(sourceStorage, sourcePath)
	if err != nil {
		return err
	}
	destFiles, err := listRecursiveSlice(destStorage, destPath)
	if err != nil {
		return err
	}
	sourceFiles = stripPathPrefix(sourceFiles, sourcePath)
	destFiles = stripPathPrefix(destFiles, sourcePath)
	sourceUnique := diff(sourceFiles, destFiles)
	destUnique := diff(destFiles, sourceFiles)

	queue := concurrency.NewWorkerQueue(context.Background(), maxWorkers)

	for _, p := range sourceUnique {
		// Variables used in closure
		sourcePath := sourcePath
		destPath := destPath
		p := p
		err := queue.Go(func() error {
			data, err := sourceStorage.Get(path.Join(sourcePath, p))
			if err != nil {
				return err
			}
			return destStorage.Put(path.Join(destPath, p), data)
		})
		if err != nil {
			return err
		}
	}
	for _, p := range destUnique {
		// Variables used in closure
		destPath := destPath
		p := p
		err := queue.Go(func() error {
			return destStorage.Delete(path.Join(destPath, p))
		})
		if err != nil {
			return err
		}
	}

	return queue.Wait()
}

// diff returns the elements in `a` that aren't in `b`.
func diff(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func stripPathPrefix(paths []string, prefix string) []string {
	ret := make([]string, len(paths))
	for i, v := range paths {
		ret[i] = strings.TrimPrefix(v, prefix)
	}
	return ret
}

func listRecursiveSlice(storage Storage, path string) ([]string, error) {
	results := make(chan ListResult)
	out := []string{}
	go storage.ListRecursive(results, path)
	for result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		out = append(out, result.Path)
	}
	return out, nil
}
