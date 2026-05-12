package application

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"

	"github.com/andrey-kalistratov/task-manager/worker/internal/config"
	"github.com/andrey-kalistratov/task-manager/worker/internal/task"
	"github.com/andrey-kalistratov/task-manager/worker/internal/task/docker"
	"github.com/andrey-kalistratov/task-manager/worker/internal/task/fs"
	"github.com/andrey-kalistratov/task-manager/worker/internal/task/kafka"
	"github.com/andrey-kalistratov/task-manager/worker/internal/task/object"
)

type App struct {
	cfg       *config.Config
	logger    *slog.Logger
	consumer  *kafka.Consumer
	resources []resource
}

type resource struct {
	Name string
	io.Closer
}

func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
	a := &App{
		cfg:    cfg,
		logger: logger,
	}

	fsStorage := fs.NewStorage(cfg.Execution.WorkDir, logger.With("component", "fs"))

	s3Storage := object.NewStorage(&cfg.Storage.S3)

	runner, err := docker.NewRunner(logger.With("component", "docker"))
	if err != nil {
		a.cleanup()
		return nil, fmt.Errorf("init docker runner: %w", err)
	}
	a.resources = append(a.resources, resource{
		Name:   "docker",
		Closer: runner,
	})

	producer := kafka.NewProducer(&cfg.Messaging)
	a.resources = append(a.resources, resource{
		Name:   "kafka producer",
		Closer: producer,
	})

	service := task.NewService(task.ServiceOptions{
		Cfg:       &cfg.Execution,
		FS:        fsStorage,
		S3:        s3Storage,
		Executor:  runner,
		Publisher: producer,
		Logger:    logger.With("component", "service"),
	})

	a.consumer = kafka.NewConsumer(&cfg.Messaging, service, logger.With("component", "kafka consumer"))
	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.cleanup()

	if err := a.consumer.Run(ctx); err != nil && ctx.Err() == nil {
		return fmt.Errorf("run consumer: %w", err)
	}
	return nil
}

func (a *App) cleanup() {
	for _, r := range slices.Backward(a.resources) {
		if err := r.Close(); err != nil {
			a.logger.Error("failed to close resource", "error", err, "resource", r.Name)
		}
	}
}
