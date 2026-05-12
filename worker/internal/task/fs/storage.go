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

	"github.com/andrey-kalistratov/task-manager/worker/internal/task"
)

var _ task.FileRepository = (*Storage)(nil)

type Storage struct {
	root   string
	logger *slog.Logger
}

func NewStorage(root string, logger *slog.Logger) *Storage {
	return &Storage{
		root:   root,
		logger: logger,
	}
}

func (s *Storage) Download(_ context.Context, path string) (io.ReadCloser, error) {
	path = filepath.Join(s.root, path)

	src, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, task.ErrFileNotFound
		}
		return nil, fmt.Errorf("open file: %w", err)
	}
	return src, nil
}

func (s *Storage) Upload(_ context.Context, path string, r io.Reader) error {
	path = filepath.Join(s.root, path)

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create parent directories for file: %w", err)
	}

	dst, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
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

func (s *Storage) List(ctx context.Context, prefix string) ([]string, error) {
	var paths []string
	root := filepath.Join(s.root, prefix)
	collectPaths := func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err != nil || d.IsDir() {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		paths = append(paths, rel)
		return nil
	}
	if err := filepath.WalkDir(root, collectPaths); err != nil {
		return nil, fmt.Errorf("collect paths: %w", err)
	}
	return paths, nil
}

func (s *Storage) Delete(_ context.Context, path string) error {
	return os.Remove(filepath.Join(s.root, path))
}
