package filelogger

import (
	"log/slog"
	"os"
	"path/filepath"
)

func New(path string, opts *slog.HandlerOptions) (*slog.Logger, func(), error) {
	f, err := openFile(path)
	if err != nil {
		return nil, func() {}, err
	}

	logger := slog.New(slog.NewTextHandler(f, opts))
	return logger, func() { f.Close() }, nil
}

func openFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}
