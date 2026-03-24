package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator"
)

// Load looks for a config file first by the given path, then by standard paths.
//
// Pass "" if no path is provided.
func Load(path string) (*Config, error) {
	var cfg Config
	if err := defaults.Set(&cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefaults, err)
	}

	if p := find(path); p != "" {
		if _, err := toml.DecodeFile(p, &cfg); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrLoad, err)
		}
	}

	if err := validator.New().Struct(&cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidate, err)
	}
	return &cfg, nil
}

func find(path string) string {
	paths := []string{
		path,
		"/etc/task-manager/planner/config.toml",
		"./planner/config/config.toml",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
