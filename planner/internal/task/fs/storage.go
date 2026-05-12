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

	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

var _ task.FileStorage = (*FileStorage)(nil)

type FileStorage struct {
	logger *slog.Logger
}

func NewStorage(logger *slog.Logger) *FileStorage {
	return &FileStorage{logger: logger}
}

func (s FileStorage) Download(_ context.Context, path string) (io.ReadCloser, error) {
	src, err := os.Open(path)
	if err != nil {
		return nil, toFileErr(path, err)
	}
	return src, nil
}

func (s FileStorage) Upload(_ context.Context, path string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return toFileErr(path, err)
	}

	dst, err := os.Create(path)
	if err != nil {
		return toFileErr(path, err)
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

func toFileErr(path string, err error) error {
	file := task.File{
		Path:     path,
		Provider: task.ProviderFS,
	}

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return &task.FileError{
			Kind: task.FileErrorNotFound,
			File: file,
			Err:  err,
		}
	case errors.Is(err, fs.ErrPermission):
		return &task.FileError{
			Kind: task.FileErrorPermission,
			File: file,
			Err:  err,
		}
	default:
		return err
	}
}
