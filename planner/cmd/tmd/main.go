package main

import (
	"flag"
	"log"

	"task-manager/planner/internal/config"
	"task-manager/planner/internal/daemon"
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
