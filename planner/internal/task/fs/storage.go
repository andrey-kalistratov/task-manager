package fs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/andrey-kalistratov/task-manager/planner/internal/oops"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

var _ task.FileStorage = (*FileStorage)(nil)

type FileStorage struct {
	logger *slog.Logger
}

func NewFileStorage(logger *slog.Logger) *FileStorage {
	return &FileStorage{logger: logger}
}

func (s FileStorage) Download(_ context.Context, path string) (io.ReadCloser, error) {
	src, err := os.Open(path)
	if err != nil {
		return nil, makeFileErr(path, err)
	}
	return src, nil
}

func (s FileStorage) Upload(_ context.Context, path string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return makeFileErr(path, err)
	}

	dst, err := os.Create(path)
	if err != nil {
		return makeFileErr(path, err)
	}
	defer func() {
		if err = dst.Close(); err != nil {
			s.logger.Error("failed to close file", "error", err, "file", dst)
		}
	}()

	_, err = io.Copy(dst, r)
	if err != nil {
		return fmt.Errorf("copy data: %w", err)
	}

	return nil
}

func makeFileErr(path string, err error) oops.FileError {
	fileErr := oops.FileError{
		File: task.File{
			Path:     path,
			Provider: task.ProviderFS,
		},
		Err: err,
	}
	switch {
	case errors.Is(err, fs.ErrNotExist):
		fileErr.Kind = oops.FileErrorNotFound
	case errors.Is(err, fs.ErrPermission):
		fileErr.Kind = oops.FileErrorPermission
	default:
		fileErr.Kind = oops.FileErrorUnknown
	}
	return fileErr
}
