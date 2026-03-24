package app

import (
	"task-manager/planner/internal/config"
	"task-manager/planner/internal/task/domain/use_cases"
)

// App represents a Planner service instance.
type App struct {
	AddUC  *use_cases.Add
	logger *Logger
}

// New creates a new App.
func New() (*App, error) {
	cfg, err := config.Load("")
	if err != nil {
		return nil, err
	}

	logger := newLogger(cfg)

	return &App{
		AddUC:  use_cases.NewAdd(logger.With("cmd", "add")),
		logger: logger,
	}, nil
}

// Close releases resources held by the App.
func (a *App) Close() error {
	return a.logger.close()
}
