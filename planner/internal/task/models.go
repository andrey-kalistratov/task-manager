package task

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID        uuid.UUID
	Status    Status
	CreatedAt time.Time

	Command string
	Name    string
	Image   string

	Inputs    map[Parameter]File
	Uploads   map[Parameter]File
	Downloads map[Parameter]File
	Outputs   map[Parameter]File
}

type Result struct {
	TaskID    uuid.UUID
	Status    Status
	Downloads map[Parameter]File
}

type Status int

const (
	StatusRunning Status = iota
	StatusSucceeded
	StatusFailed
)

func (s Status) String() string {
	switch s {
	case StatusRunning:
		return "running"
	case StatusSucceeded:
		return "succeeded"
	case StatusFailed:
		return "failed"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

type Parameter string

type File struct {
	Path     string
	Provider StorageProvider
}

func (f File) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("path", f.Path),
		slog.Any("provider", f.Provider),
	)
}

type StorageProvider int

const (
	ProviderFS StorageProvider = iota
	ProviderS3
)

func (p StorageProvider) String() string {
	switch p {
	case ProviderFS:
		return "fs"
	case ProviderS3:
		return "s3"
	default:
		return fmt.Sprintf("unknown(%d)", int(p))
	}
}
