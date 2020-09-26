package shared

import (
	"fmt"

	"replicate.ai/cli/pkg/storage"
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

func (GCSStorage) PutPath(args PutPathArgs, _ *int) error {
	st, err := storage.NewGCSStorage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.PutPath(args.Src, args.Dest)
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

func (S3Storage) PutPath(args PutPathArgs, _ *int) error {
	st, err := storage.NewS3Storage(args.Bucket, args.Root)
	if err != nil {
		return err
	}
	return st.PutPath(args.Src, args.Dest)
}

type DiskStorage struct{}

func (DiskStorage) PutPath(args PutPathArgs, _ *int) error {
	st, err := storage.NewDiskStorage(args.Root)
	if err != nil {
		return err
	}
	return st.PutPath(args.Src, args.Dest)
}
