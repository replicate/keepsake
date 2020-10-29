package shared

import (
	"fmt"

	"github.com/replicate/replicate/go/pkg/repository"
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

type GCSRepository struct{}

func (GCSRepository) Get(args GetArgs, ret *GetReturn) error {
	st, err := repository.NewGCSRepository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	ret.Data, err = st.Get(args.Path)
	// HACK: net/rpc/jsonrpc doesn't let us include error codes, so prefix with
	// predictable error name
	if _, ok := err.(*repository.DoesNotExistError); ok {
		return fmt.Errorf("DoesNotExistError:: %w", err)
	}
	return err
}

func (GCSRepository) Put(args PutArgs, _ *int) error {
	st, err := repository.NewGCSRepository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.Put(args.Path, args.Data)
}

func (GCSRepository) List(args ListArgs, ret *ListReturn) error {
	st, err := repository.NewGCSRepository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	ret.Paths, err = st.List(args.Path)
	return err
}

func (GCSRepository) PutPath(args PutPathArgs, _ *int) error {
	st, err := repository.NewGCSRepository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.PutPath(args.Src, args.Dest)
}

func (GCSRepository) PutPathTar(args PutPathTarArgs, _ *int) error {
	st, err := repository.NewGCSRepository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.PutPathTar(args.LocalPath, args.TarPath, args.IncludePath)
}

func (GCSRepository) Delete(args DeleteArgs, _ *int) error {
	st, err := repository.NewGCSRepository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.Delete(args.Path)
}

func (GCSRepository) GetPathTar(args GetPathTarArgs, _ *int) error {
	st, err := repository.NewGCSRepository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	err = st.GetPathTar(args.TarPath, args.LocalPath)
	// HACK: net/rpc/jsonrpc doesn't let us include error codes, so prefix with
	// predictable error name
	if _, ok := err.(*repository.DoesNotExistError); ok {
		return fmt.Errorf("DoesNotExistError:: %w", err)
	}
	return err
}

type S3Repository struct{}

func (S3Repository) Get(args GetArgs, ret *GetReturn) error {
	st, err := repository.NewS3Repository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	ret.Data, err = st.Get(args.Path)
	// HACK: net/rpc/jsonrpc doesn't let us include error codes, so prefix with
	// predictable error name
	if _, ok := err.(*repository.DoesNotExistError); ok {
		return fmt.Errorf("DoesNotExistError:: %w", err)
	}
	return err
}

func (S3Repository) Put(args PutArgs, _ *int) error {
	st, err := repository.NewS3Repository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.Put(args.Path, args.Data)
}

func (S3Repository) List(args ListArgs, ret *ListReturn) error {
	st, err := repository.NewS3Repository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	ret.Paths, err = st.List(args.Path)
	return err
}

func (S3Repository) PutPath(args PutPathArgs, _ *int) error {
	st, err := repository.NewS3Repository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.PutPath(args.Src, args.Dest)
}

func (S3Repository) PutPathTar(args PutPathTarArgs, _ *int) error {
	st, err := repository.NewS3Repository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.PutPathTar(args.LocalPath, args.TarPath, args.IncludePath)
}

func (S3Repository) Delete(args DeleteArgs, _ *int) error {
	st, err := repository.NewS3Repository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.Delete(args.Path)
}

func (S3Repository) GetPathTar(args GetPathTarArgs, _ *int) error {
	st, err := repository.NewS3Repository(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	err = st.GetPathTar(args.TarPath, args.LocalPath)
	// HACK: net/rpc/jsonrpc doesn't let us include error codes, so prefix with
	// predictable error name
	if _, ok := err.(*repository.DoesNotExistError); ok {
		return fmt.Errorf("DoesNotExistError:: %w", err)
	}
	return err
}

type DiskRepository struct{}

func (DiskRepository) PutPath(args PutPathArgs, _ *int) error {
	st, err := repository.NewDiskRepository(args.Root)
	if err != nil {
		return err
	}
	return st.PutPath(args.Src, args.Dest)
}

func (DiskRepository) PutPathTar(args PutPathTarArgs, _ *int) error {
	st, err := repository.NewDiskRepository(args.Root)
	if err != nil {
		return err
	}
	return st.PutPathTar(args.LocalPath, args.TarPath, args.IncludePath)
}

func (DiskRepository) Delete(args DeleteArgs, _ *int) error {
	st, err := repository.NewDiskRepository(args.Root)
	if err != nil {
		return err
	}
	return st.Delete(args.Path)
}

func (DiskRepository) GetPathTar(args GetPathTarArgs, _ *int) error {
	st, err := repository.NewDiskRepository(args.Root)
	if err != nil {
		return err
	}
	err = st.GetPathTar(args.TarPath, args.LocalPath)
	// HACK: net/rpc/jsonrpc doesn't let us include error codes, so prefix with
	// predictable error name
	if _, ok := err.(*repository.DoesNotExistError); ok {
		return fmt.Errorf("DoesNotExistError:: %w", err)
	}
	return err
}
