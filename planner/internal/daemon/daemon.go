package daemon

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/fs"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/ipc"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/kafka"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/object"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/sqlite"
)

type Daemon struct {
	cfg       *config.Config
	logger    *slog.Logger
	consumer  *kafka.Consumer
	server    *ipc.Server
	resources []resource
}

type resource struct {
	Name string
	io.Closer
}

func New(cfg *config.Config, logger *slog.Logger) (*Daemon, error) {
	d := &Daemon{
		cfg:    cfg,
		logger: logger,
	}

	sqlStorage, err := sqlite.NewStorage(&cfg.Storage, logger.With("component", "sqlite"))
	if err != nil {
		d.cleanup()
		return nil, fmt.Errorf("init storage: %w", err)
	}
	d.resources = append(d.resources, resource{
		Name:   "sqlite",
		Closer: sqlStorage,
	})

	fsStorage := fs.NewStorage(logger.With("component", "fs"))

	s3Storage := object.NewStorage(&cfg.Storage.S3)

	producer := kafka.NewProducer(&cfg.Messaging)
	d.resources = append(d.resources, resource{
		Name:   "kafka producer",
		Closer: producer,
	})

	service := task.NewService(task.ServiceOptions{
		Storage:   sqlStorage,
		FS:        fsStorage,
		S3:        s3Storage,
		Publisher: producer,
		Logger:    logger.With("component", "service"),
	})

	d.consumer = kafka.NewConsumer(&cfg.Messaging, service, logger.With("component", "kafka consumer"))

	d.server, err = ipc.NewServer(service, logger.With("component", "ipc server"))
	if err != nil {
		d.cleanup()
		return nil, fmt.Errorf("init ipc server: %w", err)
	}

	return d, nil
}

func (d *Daemon) Run(ctx context.Context) error {
	defer d.cleanup()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return d.consumer.Run(ctx)
	})
	g.Go(func() error {
		return d.server.Serve()
	})

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.cfg.ShutdownTimeout))
	defer cancel()

	if err := d.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown gracefully: %w", err)
	}

	return g.Wait()
}

func (d *Daemon) cleanup() {
	for _, r := range slices.Backward(d.resources) {
		if err := r.Close(); err != nil {
			d.logger.Error("failed to close resource", "error", err, "resource", r.Name)
		}
	}
}
