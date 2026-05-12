package task

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/uuid"
)

type Storage interface {
	Save(ctx context.Context, task *Task) error
	Get(ctx context.Context, id uuid.UUID) (*Task, error)
}

type FileStorage interface {
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	Upload(ctx context.Context, path string, r io.Reader) error
}

type Publisher interface {
	Publish(ctx context.Context, task *Task) error
}

type Service struct {
	storage   Storage
	fs        FileStorage
	s3        FileStorage
	publisher Publisher
	logger    *slog.Logger
}

type ServiceOptions struct {
	Storage   Storage
	FS        FileStorage
	S3        FileStorage
	Publisher Publisher
	Logger    *slog.Logger
}

func NewService(opts ServiceOptions) *Service {
	return &Service{
		storage:   opts.Storage,
		fs:        opts.FS,
		s3:        opts.S3,
		publisher: opts.Publisher,
		logger:    opts.Logger,
	}
}

func (s *Service) Run(ctx context.Context, t *Task) error {
	t.Uploads = make(map[string]File, len(t.Inputs))
	for path, input := range t.Inputs {
		upload := File{
			Path:     fmt.Sprintf("%s/%s", t.ID, path),
			Provider: ProviderS3,
		}
		if err := s.transferFile(ctx, input, upload); err != nil {
			return fmt.Errorf("transfer input file from user: %w", err)
		}

		t.Uploads[path] = upload
	}

	t.Downloads = make(map[string]File, len(t.Outputs))
	for path := range t.Outputs {
		t.Downloads[path] = File{
			Path:     fmt.Sprintf("%s/%s", t.ID, path),
			Provider: ProviderS3,
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

func (s *Service) Record(ctx context.Context, result *Result) error {
	t, err := s.storage.Get(ctx, result.TaskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	t.Status = result.Status

	for param, output := range t.Outputs {
		if err = s.transferFile(ctx, t.Downloads[param], output); err != nil {
			if fileErr, ok := errors.AsType[*FileError](err); ok && fileErr.Kind == FileErrorNotFound {
				s.logger.Info("output file was not created", "task", t.ID, "param", param)
				continue
			}
			return fmt.Errorf("transfer output file from user: %w", err)
		}
	}

	if err = s.storage.Save(ctx, t); err != nil {
		return fmt.Errorf("save task: %w", err)
	}
	return nil
}

func (s *Service) transferFile(ctx context.Context, src, dst File) error {
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

func (s *Service) downloadFile(ctx context.Context, f File) (io.ReadCloser, error) {
	switch f.Provider {
	case ProviderFS:
		return s.fs.Download(ctx, f.Path)
	case ProviderS3:
		return s.s3.Download(ctx, f.Path)
	default:
		return nil, fmt.Errorf("unknown storage provider: %v", f.Provider)
	}
}

func (s *Service) uploadFile(ctx context.Context, f File, r io.Reader) error {
	switch f.Provider {
	case ProviderFS:
		return s.fs.Upload(ctx, f.Path, r)
	case ProviderS3:
		return s.s3.Upload(ctx, f.Path, r)
	default:
		return fmt.Errorf("unknown storage provider: %v", f.Provider)
	}
}
