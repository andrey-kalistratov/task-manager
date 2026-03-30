package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
)

const UnixSocket = "/var/run/planner.sock"

type (
	// Config represents the application configuration.
	Config struct {
		Logging LogConfig `json:"logging"`
		DB      DBConfig  `json:"db"`
	}

	// LogConfig contains logging configuration for the application.
	LogConfig struct {
		// Level sets the logging level. Possible values: debug, info, warn, error.
		Level slog.Level `json:"level"`

		// File specifies the path to the log file.
		File string `json:"file"`
	}

	// DBConfig contains database configuration for the application.
	DBConfig struct {
		// File specifies the path to the database file.
		File string `json:"file"`
	}
)

var (
	// ErrFind indicates that config was not found neither by the given path nor by standard ones.
	ErrFind = errors.New("failed to find config")

	// ErrLoad indicates that loading or parsing the config file failed.
	ErrLoad = errors.New("failed to load config")
)

// Load looks for a config file first by the given path, then by standard paths.
//
// Pass "" if no path is provided.
func Load(path string) (*Config, error) {
	cfg := newDefault()

	paths := []string{
		path,
		"/etc/task-manager/planner/config.json",
		"./planner/internal/config/testdata/simple.json",
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrLoad, err)
		}
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrLoad, err)
		}
		return cfg, nil
	}
	return nil, ErrFind
}

func newDefault() *Config {
	return &Config{
		Logging: LogConfig{
			Level: slog.LevelError,
			File:  "/var/log/task-manager/planner.log",
		},
		DB: DBConfig{
			File: "/var/lib/task-manager/planner.db",
		},
	}
}
