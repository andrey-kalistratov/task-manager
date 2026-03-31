package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/andrey-kalistratov/task-manager/worker/internal/config"
	"github.com/andrey-kalistratov/task-manager/worker/internal/pool"
	"github.com/andrey-kalistratov/task-manager/worker/internal/queue"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Info("unable to load config", "err", err)
		return
	}
	slog.Info("config", "cfg", cfg)

	consumer := queue.NewKafka(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID)
	p := pool.New(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	if err := p.Run(ctx, consumer); err != nil {
		slog.Error("finished with error!", "err", err)
	}

	slog.Info("Exitting!")
}
