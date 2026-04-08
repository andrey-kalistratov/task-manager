package main

import (
	"flag"
	"log"

	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
	"github.com/andrey-kalistratov/task-manager/planner/internal/daemon"
)

func main() {
	cfgPath := flag.String("config", "", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v\n", err)
	}

	if err := daemon.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
