package main

import "C"

import (
	"replicate.ai/cli/pkg/storage"
)

//export DiskStorageGet
func DiskStorageGet(rootDir string, path string, result *[]C.char) {
	storage, err := storage.NewDiskStorage(rootDir)
	if err != nil {
		panic(err)
	}
	r, err := storage.Get(path)
	if err != nil {
		panic(err)
	}
	for _, b := range r {
		*result = append(*result, C.char(b))
	}
}

func main() {}
