package daemon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/broker"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/business"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/fs"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/ipc"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/object"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/sqlite"
)

type Daemon struct {
	cfg        *config.Config
	logger     *slog.Logger
	transports []transport
	resources  []resource
}

type transport struct {
	Name string
	task.Transport
}

type resource struct {
	Name string
	io.Closer
}

func New(cfg *config.Config, logger *slog.Logger) (d *Daemon, err error) {
	d = &Daemon{
		cfg:    cfg,
		logger: logger,
	}

	defer func() {
		if err != nil {
			d.releaseResources()
		}
	}()

	storage, err := sqlite.NewStorage(cfg, logger)
	if err != nil {
		return d, fmt.Errorf("init storage: %w", err)
	}
	d.resources = append(d.resources, resource{
		Name:   "storage",
		Closer: storage,
	})

	fsStorage := fs.NewFileStorage()

	s3Storage := object.NewFileStorage(cfg)

	publisher := broker.NewProducer(cfg)
	d.resources = append(d.resources, resource{
		Name:   "publisher",
		Closer: publisher,
	})

	service := business.NewService(business.Options{
		Storage:   storage,
		FSStorage: fsStorage,
		S3Storage: s3Storage,
		Publisher: publisher,
		Logger:    logger.With("component", "service"),
	})

	consumer := broker.NewResultConsumer(cfg, service, logger.With("component", "consumer"))
	d.transports = append(d.transports, transport{
		Name:      "consumer",
		Transport: consumer,
	})

	server := ipc.NewServer(service, logger.With("component", "server"))
	d.transports = append(d.transports, transport{
		Name:      "server",
		Transport: server,
	})

	return
}

func (d *Daemon) Run(ctx context.Context) error {
	defer d.releaseResources()

	g, ctx := errgroup.WithContext(ctx)

	for _, t := range d.transports {
		g.Go(func() error {
			return t.Serve(ctx)
		})
	}

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.cfg.ShutdownTimeout))
	defer cancel()

	d.shutdownTransport(ctx)

	errc := make(chan error, 1)
	go func() {
		errc <- g.Wait()
	}()

	var err error
	select {
	case err = <-errc:
	case <-ctx.Done():
		err = errors.New("graceful shutdown timed out")
	}
	return err
}

func (d *Daemon) shutdownTransport(ctx context.Context) {
	var wg sync.WaitGroup
	for _, t := range d.transports {
		wg.Go(func() {
			if err := t.Shutdown(ctx); err != nil {
				d.logger.Error("failed to shutdown transport", "transport", t.Name, "error", err)
			}
		})
	}
	wg.Wait()
}

func (d *Daemon) releaseResources() {
	for _, r := range slices.Backward(d.resources) {
		if err := r.Close(); err != nil {
			d.logger.Error("failed to close resource", "resource", r.Name, "error", err)
		}
	}
}
