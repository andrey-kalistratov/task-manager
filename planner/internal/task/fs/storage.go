package fs

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/andrey-kalistratov/task-manager/planner/internal/oops"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

var _ task.FileStorage = (*FileStorage)(nil)

type FileStorage struct{}

func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

func (s FileStorage) Download(_ context.Context, file task.File) (io.ReadCloser, error) {
	src, err := os.Open(file.Path)
	switch {
	case os.IsNotExist(err):
		return nil, oops.FileError{
			Kind: oops.FileErrorNotFound,
			File: file,
		}
	case os.IsPermission(err):
		return nil, oops.FileError{
			Kind: oops.FileErrorPermission,
			File: file,
		}
	case err != nil:
		return nil, fmt.Errorf("open file: %w", err)
	}
	return src, nil
}

func (s FileStorage) Upload(ctx context.Context, path string, r io.Reader) (task.File, error) {
	//TODO implement me
	panic("implement me")
}
