package config

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSimpleConfig(t *testing.T) {
	cfg, err := Load("testdata/simple.json")
	if err != nil {
		t.Fatal(err)
	}

	expected := &Config{
		Logging: LogConfig{
			Level: slog.LevelDebug,
			File:  "/var/log/task-manager/planner.log",
		},
	}
	if diff := cmp.Diff(expected, cfg); diff != "" {
		t.Errorf("Load() mismatch (-want +got):\n%s", diff)
	}
}

func TestCorruptedConfig(t *testing.T) {
	if _, err := Load("testdata/corrupted.json"); !errors.Is(err, ErrLoad) {
		t.Error("expected ErrLoad")
	}
}

func TestInvalidLogLevel(t *testing.T) {
	if _, err := Load("testdata/invalid_log_level.json"); !errors.Is(err, ErrLoad) {
		t.Error("expected ErrValidate")
	}
}
