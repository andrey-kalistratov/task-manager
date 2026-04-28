package business

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

var _ task.Service = (*Service)(nil)

type Service struct {
	storage   task.Storage
	fsStorage task.FileStorage
	s3Storage task.FileStorage
	publisher task.Publisher
	logger    *slog.Logger
}

type Options struct {
	Storage   task.Storage
	FSStorage task.FileStorage
	S3Storage task.FileStorage
	Publisher task.Publisher
	Logger    *slog.Logger
}

func NewService(opts Options) *Service {
	return &Service{
		storage:   opts.Storage,
		fsStorage: opts.FSStorage,
		s3Storage: opts.S3Storage,
		publisher: opts.Publisher,
		logger:    opts.Logger,
	}
}

func (s *Service) Run(ctx context.Context, t *task.Task) error {
	t.Uploads = make(map[task.Parameter]task.File, len(t.Inputs))
	for param, input := range t.Inputs {
		r, err := s.fsStorage.Download(ctx, input)
		if err != nil {
			return fmt.Errorf("download input %q: %w", param, err)
		}

		upload, err := s.s3Storage.Upload(ctx, r)
		if err != nil {
			return fmt.Errorf("upload input %q: %w", param, err)
		}
		t.Uploads[param] = upload

		if err = r.Close(); err != nil {
			s.logger.Error(
				"close file", "error", err, "path", input.Path, "provider", input.Provider,
			)
		}
	}

	if err := s.publisher.Publish(ctx, t); err != nil {
		return fmt.Errorf("publish task: %w", err)
	}

	if err := s.storage.Save(ctx, t); err != nil {
		return fmt.Errorf("save task: %w", err)
	}
	return nil
}

func (s *Service) Record(ctx context.Context, result task.Result) error {
	t, err := s.storage.Get(ctx, result.TaskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	t.Status = result.Status

	if err = s.storage.Save(ctx, t); err != nil {
		return fmt.Errorf("save task: %w", err)
	}
	return nil
}
