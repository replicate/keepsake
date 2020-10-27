package shared

import (
	"fmt"

	"github.com/replicate/replicate/go/pkg/storage"
)

type GetArgs struct {
	Bucket, Root, Path string
}

type GetReturn struct {
	Data []byte
}

type PutArgs struct {
	Bucket, Root, Path string
	Data               []byte
}

type PutPathArgs struct {
	Bucket, Root, Src, Dest string
}

type PutPathTarArgs struct {
	Bucket, Root, LocalPath, TarPath, IncludePath string
}

type ListArgs struct {
	Bucket, Root, Path string
}
type ListReturn struct {
	Paths []string
}

type DeleteArgs struct {
	Bucket, Root, Path string
}

type GetPathTarArgs struct {
	Bucket, Root, TarPath, LocalPath string
}

type GCSStorage struct{}

func (GCSStorage) Get(args GetArgs, ret *GetReturn) error {
	st, err := storage.NewGCSStorage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	ret.Data, err = st.Get(args.Path)
	// HACK: net/rpc/jsonrpc doesn't let us include error codes, so prefix with
	// predictable error name
	if _, ok := err.(*storage.DoesNotExistError); ok {
		return fmt.Errorf("DoesNotExistError:: %w", err)
	}
	return err
}

func (GCSStorage) Put(args PutArgs, _ *int) error {
	st, err := storage.NewGCSStorage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.Put(args.Path, args.Data)
}

func (GCSStorage) List(args ListArgs, ret *ListReturn) error {
	st, err := storage.NewGCSStorage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	ret.Paths, err = st.List(args.Path)
	return err
}

func (GCSStorage) PutPath(args PutPathArgs, _ *int) error {
	st, err := storage.NewGCSStorage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.PutPath(args.Src, args.Dest)
}

func (GCSStorage) PutPathTar(args PutPathTarArgs, _ *int) error {
	st, err := storage.NewGCSStorage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.PutPathTar(args.LocalPath, args.TarPath, args.IncludePath)
}

func (GCSStorage) Delete(args DeleteArgs, _ *int) error {
	st, err := storage.NewGCSStorage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.Delete(args.Path)
}

func (GCSStorage) GetPathTar(args GetPathTarArgs, _ *int) error {
	st, err := storage.NewGCSStorage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	err = st.GetPathTar(args.TarPath, args.LocalPath)
	// HACK: net/rpc/jsonrpc doesn't let us include error codes, so prefix with
	// predictable error name
	if _, ok := err.(*storage.DoesNotExistError); ok {
		return fmt.Errorf("DoesNotExistError:: %w", err)
	}
	return err
}

type S3Storage struct{}

func (S3Storage) Get(args GetArgs, ret *GetReturn) error {
	st, err := storage.NewS3Storage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	ret.Data, err = st.Get(args.Path)
	// HACK: net/rpc/jsonrpc doesn't let us include error codes, so prefix with
	// predictable error name
	if _, ok := err.(*storage.DoesNotExistError); ok {
		return fmt.Errorf("DoesNotExistError:: %w", err)
	}
	return err
}

func (S3Storage) Put(args PutArgs, _ *int) error {
	st, err := storage.NewS3Storage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.Put(args.Path, args.Data)
}

func (S3Storage) List(args ListArgs, ret *ListReturn) error {
	st, err := storage.NewS3Storage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	ret.Paths, err = st.List(args.Path)
	return err
}

func (S3Storage) PutPath(args PutPathArgs, _ *int) error {
	st, err := storage.NewS3Storage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.PutPath(args.Src, args.Dest)
}

func (S3Storage) PutPathTar(args PutPathTarArgs, _ *int) error {
	st, err := storage.NewS3Storage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.PutPathTar(args.LocalPath, args.TarPath, args.IncludePath)
}

func (S3Storage) Delete(args DeleteArgs, _ *int) error {
	st, err := storage.NewS3Storage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.Delete(args.Path)
}

func (S3Storage) GetPathTar(args GetPathTarArgs, _ *int) error {
	st, err := storage.NewS3Storage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	err = st.GetPathTar(args.TarPath, args.LocalPath)
	// HACK: net/rpc/jsonrpc doesn't let us include error codes, so prefix with
	// predictable error name
	if _, ok := err.(*storage.DoesNotExistError); ok {
		return fmt.Errorf("DoesNotExistError:: %w", err)
	}
	return err
}

type DiskStorage struct{}

func (DiskStorage) PutPath(args PutPathArgs, _ *int) error {
	st, err := storage.NewDiskStorage(args.Root)
	if err != nil {
		return err
	}
	return st.PutPath(args.Src, args.Dest)
}

func (DiskStorage) PutPathTar(args PutPathTarArgs, _ *int) error {
	st, err := storage.NewDiskStorage(args.Root)
	if err != nil {
		return err
	}
	return st.PutPathTar(args.LocalPath, args.TarPath, args.IncludePath)
}

func (DiskStorage) Delete(args DeleteArgs, _ *int) error {
	st, err := storage.NewDiskStorage(args.Root)
	if err != nil {
		return err
	}
	return st.Delete(args.Path)
}

func (DiskStorage) GetPathTar(args GetPathTarArgs, _ *int) error {
	st, err := storage.NewDiskStorage(args.Root)
	if err != nil {
		return err
	}
	err = st.GetPathTar(args.TarPath, args.LocalPath)
	// HACK: net/rpc/jsonrpc doesn't let us include error codes, so prefix with
	// predictable error name
	if _, ok := err.(*storage.DoesNotExistError); ok {
		return fmt.Errorf("DoesNotExistError:: %w", err)
	}
	return err
}
