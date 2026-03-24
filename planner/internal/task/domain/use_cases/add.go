package use_cases

import (
	"log/slog"
)

// Add handles adding a task to the execution graph.
type Add struct {
	logger *slog.Logger
}

// NewAdd creates a new Add use case.
func NewAdd(logger *slog.Logger) *Add {
	return &Add{logger: logger}
}

// Execute runs the Add use case.
func (a *Add) Execute() error {
	a.logger.Warn("dummy implementation")
	return nil
}
