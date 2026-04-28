package config

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestLoad(t *testing.T) {
	for _, tc := range []struct {
		name     string
		file     string
		expected *Config
		err      error
	}{
		{
			name: "simple",
			file: "testdata/simple.json",
			expected: &Config{
				LogLevel: slog.LevelDebug,
				Storage: StorageConfig{
					SqliteFile: "task-manager/planner.db",
				},
				ShutdownTimeout: Duration(3 * time.Second),
			},
			err: nil,
		},
		{
			name: "default",
			file: "",
			expected: &Config{
				LogLevel: slog.LevelError,
				Storage: StorageConfig{
					SqliteFile: "/var/lib/task-manager/planner.db",
				},
				ShutdownTimeout: Duration(5 * time.Second),
			},
			err: nil,
		},
		{
			name:     "corrupted",
			file:     "testdata/corrupted.json",
			expected: nil,
			err:      ErrLoad,
		},
		{
			name:     "invalid",
			file:     "testdata/invalid.json",
			expected: nil,
			err:      ErrLoad,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := Load(tc.file)
			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Errorf("Load() error = %v, expected error = %v", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.expected, cfg); diff != "" {
				t.Errorf("Load() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
