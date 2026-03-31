package config

import (
	"encoding/json"
	"flag"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	WorkerCount int    `json:"worker_count"`
	LogLevel    string `json:"log_level"`
	Kafka       Kafka  `json:"kafka"`
}

type Kafka struct {
	Topic   string   `json:"topic"`
	Brokers []string `json:"brokers"`
	GroupID string   `json:"group_id"`
}

func Load() (*Config, error) {
	var cfg Config
	var path string

	flag.StringVar(&path, "config", "", "path to config file")

	flag.Parse()

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if str := strings.Fields(os.Getenv("KAFKA_BROKERS")); len(str) != 0 {
		cfg.Kafka.Brokers = str
	}

	if str := os.Getenv("KAFKA_CONSUME_TOPIC"); str != "" {
		cfg.Kafka.Topic = str
	}

	if str := os.Getenv("KAFKA_GROUP_ID"); str != "" {
		cfg.Kafka.GroupID = str
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
