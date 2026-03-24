package config

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, config string) (path string, cleanup func()) {
	f, err := os.CreateTemp("", "tm-*")
	require.NoError(t, err)

	_, err = f.WriteString(config)
	require.NoError(t, err)

	require.NoError(t, f.Close())

	cleanup = func() {
		require.NoError(t, os.Remove(f.Name()))
	}
	return f.Name(), cleanup
}

func testWithStr(t *testing.T, config string) (*Config, error) {
	cfgPath, cleanup := writeFile(t, config)
	defer cleanup()

	return Load(cfgPath)
}

func TestSimpleConfig(t *testing.T) {
	cfgStr := `[log]
	level = "debug"
	file = "/var/log/task-manager/planner.log"
	`

	cfg, err := testWithStr(t, cfgStr)
	require.NoError(t, err)

	expected := &Config{
		Log: Log{
			Level: "debug",
			File:  "/var/log/task-manager/planner.log",
		},
	}
	require.Equal(t, expected, cfg)
}

func TestCorruptedConfig(t *testing.T) {
	cfgStr := `[log
	level = "warn"
	file = "/var/log/task-manager/planner.log"
	`

	if _, err := testWithStr(t, cfgStr); !errors.Is(err, ErrLoad) {
		t.Error("expected ErrLoad")
	}
}

func TestInvalidLogLevel(t *testing.T) {
	cfgStr := `[log]
	level = "warning"
	file = "/var/log/task-manager/planner.log"
	`

	if _, err := testWithStr(t, cfgStr); !errors.Is(err, ErrValidate) {
		t.Error("expected ErrLoad")
	}
}
