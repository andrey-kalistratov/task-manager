package main

import (
	"log"
	"log/slog"

	"github.com/andrey-kalistratov/task-manager/worker/internal/config"
	"github.com/andrey-kalistratov/task-manager/worker/internal/pool"
	"github.com/andrey-kalistratov/task-manager/worker/internal/queue"
)

func main() {
	cfg, err := config.Load("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	consumer := queue.NewMock()
	p := pool.New(cfg)

	if err := p.Run(consumer); err != nil {
		log.Fatal(err)
	}

	slog.Info("Exitting!")
}
