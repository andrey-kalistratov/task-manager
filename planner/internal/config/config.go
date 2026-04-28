package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"
)

const UnixSocket = "/var/run/task-manager/planner.sock"

// Config represents the application configuration.
type Config struct {
	// LogLevel sets the logging level. Possible values: debug, info, warn, error.
	LogLevel slog.Level `json:"log_level"`

	Storage   StorageConfig `json:"storage"`
	Messaging MessageConfig `json:"messaging"`

	// ShutdownTimeout specifies how much time the application has to finish execution gracefully.
	ShutdownTimeout Duration `json:"shutdown_timeout"`
}

// StorageConfig represents storage configuration.
type StorageConfig struct {
	// SqliteFile is the path to the sqlite database file.
	SqliteFile string `json:"sqlite_file"`

	S3 S3Config `json:"s3"`
}

// S3Config represents s3 storage configuration.
type S3Config struct {
	// Endpoint is the storage server URL.
	Endpoint string `json:"endpoint"`

	// Bucket is the target bucket name. Must exist on the server.
	Bucket string `json:"bucket"`

	// Region is used for SigV4 signing. Any non-empty value works
	// for self-hosted storage if it matches the server config.
	Region string `json:"region"`

	// AccessKeyID identifies the client. Sent with every request.
	AccessKeyID string `json:"access_key_id"`

	// SecretAccessKey signs requests locally. Never sent over the wire.
	SecretAccessKey string `json:"secret_access_key"`
}

// TODO: integrate MessageConfig, S3Config

// MessageConfig represents message broker configuration.
type MessageConfig struct {
	// Brokers is the list of broker addresses (host:port).
	Brokers []string `json:"brokers"`

	Topics   TopicsConfig   `json:"topics"`
	GroupIDs GroupIDsConfig `json:"group_ids"`
}

// TopicsConfig holds broker topic names for each message type.
type TopicsConfig struct {
	// Tasks is the topic for tasks pending execution.
	Tasks string `json:"tasks"`

	// Results is the topic for executed tasks' results.
	Results string `json:"results"`
}

// GroupIDsConfig maps service names to their broker consumer group IDs.
type GroupIDsConfig struct {
	// Planner is the planner's consumer group.
	Planner string `json:"planner"`
}

// Duration is time.Duration that can be unmarshaled from JSON string.
type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(dur)
	return nil
}

// ErrLoad indicates that loading or parsing the config file failed.
var ErrLoad = errors.New("failed to load config")

// Load looks for a config file first by the given path, then by standard paths.
//
// Pass "" if no path is provided.
func Load(path string) (*Config, error) {
	cfg := newDefault()

	paths := []string{
		path,
		"/etc/task-manager/planner/config.json",
	}
	for _, path = range paths {
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrLoad, err)
		}
		if err = json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrLoad, err)
		}
		return cfg, nil
	}
	return cfg, nil
}

func newDefault() *Config {
	return &Config{
		LogLevel: slog.LevelError,
		Storage: StorageConfig{
			SqliteFile: "/var/lib/task-manager/planner.db",
		},
		Messaging: MessageConfig{
			Topics: TopicsConfig{
				Tasks:   "tasks",
				Results: "results",
			},
		},
		ShutdownTimeout: Duration(5 * time.Second),
	}
}
