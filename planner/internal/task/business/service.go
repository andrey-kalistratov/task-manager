package business

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/andrey-kalistratov/task-manager/planner/internal/oops"
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
		upload := task.File{
			Path:     fmt.Sprintf("tasks/%s/%s", t.ID, param),
			Provider: task.ProviderS3,
		}
		if err := s.transferFile(ctx, input, upload); err != nil {
			return fmt.Errorf("copy parameter %q file: %w", param, err)
		}
		t.Uploads[param] = upload
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

	t.Downloads = result.Downloads
	for param, output := range t.Outputs {
		download, ok := t.Downloads[param]
		if !ok {
			return oops.ParameterError{
				Kind:  oops.ParameterErrorOutputNotFound,
				Param: param,
			}
		}

		if err = s.transferFile(ctx, download, output); err != nil {
			return fmt.Errorf("copy parameter %q file: %w", param, err)
		}
	}

	if err = s.storage.Save(ctx, t); err != nil {
		return fmt.Errorf("save task: %w", err)
	}
	return nil
}

func (s *Service) transferFile(ctx context.Context, src, dst task.File) error {
	r, err := s.downloadFile(ctx, src)
	if err != nil {
		return fmt.Errorf("download file: %w", err)
	}
	defer func() {
		if err = r.Close(); err != nil {
			s.logger.Error("close file", "error", err, "file", src)
		}
	}()

	if err = s.uploadFile(ctx, dst, r); err != nil {
		return fmt.Errorf("upload file: %w", err)
	}
	return nil
}

func (s *Service) downloadFile(ctx context.Context, f task.File) (io.ReadCloser, error) {
	switch f.Provider {
	case task.ProviderFS:
		return s.fsStorage.Download(ctx, f.Path)
	case task.ProviderS3:
		return s.s3Storage.Download(ctx, f.Path)
	default:
		return nil, fmt.Errorf("unknown storage provider: %v", f.Provider)
	}
}

func (s *Service) uploadFile(ctx context.Context, f task.File, r io.Reader) error {
	switch f.Provider {
	case task.ProviderFS:
		return s.fsStorage.Upload(ctx, f.Path, r)
	case task.ProviderS3:
		return s.s3Storage.Upload(ctx, f.Path, r)
	default:
		return fmt.Errorf("unknown storage provider: %v", f.Provider)
	}
}
