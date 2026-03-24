package app

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"task-manager/planner/internal/config"
)

// Logger is a slog.Logger writing to a file.
//
// If given file can't be opened, writes to os.Stdout.
type Logger struct {
	*slog.Logger
	file *os.File
}

func newLogger(cfg *config.Config) *Logger {
	w, f := openLogWriter(cfg.Log.File)

	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: logLevel(cfg.Log.Level),
	})

	return &Logger{
		Logger: slog.New(handler),
		file:   f,
	}
}

func openLogWriter(path string) (io.Writer, *os.File) {
	f, err := prepareFile(path)
	if err != nil {
		fmt.Printf("logs to stdout: failed to open log file: %v\n", err)
		return os.Stdout, nil
	}
	return f, f
}

func prepareFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func logLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		panic(fmt.Sprintf("unknown log level: %s", level))
	}
}

func (l *Logger) close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
