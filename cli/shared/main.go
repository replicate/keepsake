package main

import "C"

import (
	"replicate.ai/cli/pkg/storage"
)

//export DiskStorageGet
func DiskStorageGet(rootDir string, path string, result *[]C.char, errString *string) {
	storage, err := storage.NewDiskStorage(rootDir)
	if err != nil {
		*errString = err.Error()
		return
	}
	r, err := storage.Get(path)
	if err != nil {
		*errString = err.Error()
	}
	for _, b := range r {
		*result = append(*result, C.char(b))
	}
}

func main() {}
