package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
	"github.com/andrey-kalistratov/task-manager/planner/internal/daemon"
)

// Inspired by https://grafana.com/blog/how-i-write-http-services-in-go-after-13-years.
func main() {
	if err := run(context.Background()); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfgPath := flag.String("config", "", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	logger.Info("loaded config")

	app, err := daemon.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("initialize daemon: %w", err)
	}
	logger.Info("initialized daemon")

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err = app.Run(ctx); err != nil {
		return fmt.Errorf("run daemon: %w", err)
	}
	logger.Info("daemon exited")
	return nil
}
