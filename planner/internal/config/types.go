package config

type (
	// Config represents the application configuration.
	Config struct {
		Log Log `toml:"log"`
	}

	// Log contains logging configuration for the application.
	Log struct {
		// Level sets the logging level. Possible values: debug, info, warn, error.
		Level string `toml:"level" default:"error" validate:"oneof=debug info warn error"`

		// File specifies the path to the log file.
		File string `toml:"file" default:"/var/log/task-manager/planner.log"`
	}
)
