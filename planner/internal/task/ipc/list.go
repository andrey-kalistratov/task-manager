package ipc

import (
	"log/slog"

	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

type ListHandler struct {
	service task.Service
	logger  *slog.Logger
}

func NewListHandler(service task.Service, logger *slog.Logger) *ListHandler {
	return &ListHandler{
		service: service,
		logger:  logger,
	}
}

type Task struct {
	ShortID  string `json:"short_id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Created  string `json:"created"`
	Duration string `json:"duration"`
	Command  string `json:"command"`
}

type ListResult struct {
	Tasks []Task `json:"tasks"`
}
