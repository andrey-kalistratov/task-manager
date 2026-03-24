package config

import "errors"

var (
	// ErrInvalidDefaults is returned when Config or its nested structs have invalid `default` tags.
	ErrInvalidDefaults = errors.New("invalid config default values")

	// ErrLoad is returned when loading or parsing the Config from a file fails.
	ErrLoad = errors.New("failed to load config")

	// ErrValidate is returned when the provided Config validation fails.
	ErrValidate = errors.New("invalid config")
)
