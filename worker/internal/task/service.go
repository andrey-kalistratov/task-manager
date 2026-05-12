package task

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/andrey-kalistratov/task-manager/worker/internal/config"
)

type FileStorage interface {
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	Upload(ctx context.Context, path string, r io.Reader) error
}

type FileRepository interface {
	FileStorage
	List(ctx context.Context, prefix string) ([]string, error)
	Delete(ctx context.Context, path string) error
}

type Executor interface {
	Execute(ctx context.Context, spec ExecSpec) (Execution, error)
}

type Publisher interface {
	Publish(ctx context.Context, result Result) error
}

type Service struct {
	cfg       *config.ExecutionConfig
	fs        FileRepository
	s3        FileStorage
	executor  Executor
	publisher Publisher
	logger    *slog.Logger
}

type ServiceOptions struct {
	Cfg       *config.ExecutionConfig
	FS        FileRepository
	S3        FileStorage
	Executor  Executor
	Publisher Publisher
	Logger    *slog.Logger
}

func NewService(opts ServiceOptions) *Service {
	return &Service{
		cfg:       opts.Cfg,
		fs:        opts.FS,
		s3:        opts.S3,
		executor:  opts.Executor,
		publisher: opts.Publisher,
		logger:    opts.Logger,
	}
}

var ErrFileNotFound = errors.New("file not found")

func (s *Service) Process(ctx context.Context, t *Task) error {
	for path, input := range t.Inputs {
		local := File{
			Path:     filepath.Join(t.ID.String(), path),
			Provider: ProviderTask,
		}

		if err := s.transferFile(ctx, input, local); err != nil {
			return fmt.Errorf("transfer input file to local fs: %w", err)
		}
	}

	spec := ExecSpec{
		Command: t.Command,
		Image:   s.cfg.DefaultImage,
		WorkDir: filepath.Join(s.cfg.WorkDir, t.ID.String()),
	}
	exec, err := s.executor.Execute(ctx, spec)
	if err != nil {
		return fmt.Errorf("execute task: %w", err)
	}

	for path, output := range t.Outputs {
		local := File{
			Path:     filepath.Join(t.ID.String(), path),
			Provider: ProviderTask,
		}

		if err = s.transferFile(ctx, local, output); err != nil {
			if errors.Is(err, ErrFileNotFound) {
				s.logger.Info("output file was not created", "error", err)
				continue
			}
			return fmt.Errorf("transfer output file from local fs: %w", err)
		}
	}

	result := Result{TaskID: t.ID}
	if exec.ExitCode == 0 {
		result.Status = StatusSucceeded
	} else {
		result.Status = StatusFailed
	}

	if err = s.publisher.Publish(ctx, result); err != nil {
		return fmt.Errorf("publish task: %w", err)
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
	case ProviderTask:
		return s.fs.Download(ctx, f.Path)
	case ProviderS3:
		return s.s3.Download(ctx, f.Path)
	default:
		return nil, fmt.Errorf("unknown storage provider: %v", f.Provider)
	}
}

func (s *Service) uploadFile(ctx context.Context, f File, r io.Reader) error {
	switch f.Provider {
	case ProviderTask:
		return s.fs.Upload(ctx, f.Path, r)
	case ProviderS3:
		return s.s3.Upload(ctx, f.Path, r)
	default:
		return fmt.Errorf("unknown storage provider: %v", f.Provider)
	}
}
