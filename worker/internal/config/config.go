package config

import (
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	WorkerCount int    `toml:"worker_count"`
	LogLevel    string `toml:"log_level"`
	Queue       Queue  `toml:"queue"`
}

type Queue struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

func Load(path string) (*Config, error) {
	var cfg Config

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}

	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 1
	}

	var logOpt slog.HandlerOptions

	switch cfg.LogLevel {
	case "info":
		logOpt.Level = slog.LevelInfo
	case "debug":
		logOpt.Level = slog.LevelDebug
	case "warn":
		logOpt.Level = slog.LevelWarn
	case "error":
		logOpt.Level = slog.LevelError

	default:
		logOpt.Level = slog.LevelInfo
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &logOpt)))

	return &cfg, nil
}
