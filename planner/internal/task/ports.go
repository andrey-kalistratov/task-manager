package task

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type Transport interface {
	Serve(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type Publisher interface {
	Publish(ctx context.Context, task *Task) error
}

type Service interface {
	Run(ctx context.Context, task *Task) error
	Record(ctx context.Context, result Result) error
}

type Storage interface {
	Save(ctx context.Context, task *Task) error
	Get(ctx context.Context, id uuid.UUID) (*Task, error)
}

type FileStorage interface {
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	Upload(ctx context.Context, path string, r io.Reader) error
}

//TODO указатель на file
