package ipc

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/andrey-kalistratov/task-manager/planner/internal/oops"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
	"github.com/andrey-kalistratov/task-manager/planner/unix"
)

type RunHandler struct {
	service task.Service
	logger  *slog.Logger
}

func NewRunHandler(service task.Service, logger *slog.Logger) *RunHandler {
	return &RunHandler{
		service: service,
		logger:  logger,
	}
}

type RunOptions struct {
	Command string            `json:"command"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Inputs  map[string]string `json:"inputs"`
	Outputs map[string]string `json:"outputs"`
}

type RunResult struct {
	ID uuid.UUID `json:"id"`
}

func (h *RunHandler) ServeIPC(ctx context.Context, req *unix.Request) unix.Response {
	var opts RunOptions
	if err := json.Unmarshal(req.Body, &opts); err != nil {
		h.logger.Error("failed to unmarshal run request", "error", err)
		return unix.Response{Error: "internal error"}
	}

	t := taskFromRunOptions(opts)

	if err := h.service.Run(ctx, t); err != nil {
		h.logger.Error("failed to run task", "error", err)

		if fileErr, ok := errors.AsType[oops.FileError](err); ok {
			return unix.Response{Error: fileErr.Error()}
		}
		return unix.Response{Error: "internal error"}
	}

	body, err := json.Marshal(RunResult{ID: t.ID})
	if err != nil {
		h.logger.Error("failed to marshal run result", "error", err)
		return unix.Response{Error: "internal error"}
	}

	h.logger.Info("ran task", "id", t.ID)
	return unix.Response{Body: body}
}

func taskFromRunOptions(opts RunOptions) *task.Task {
	t := &task.Task{
		ID:        uuid.New(),
		Status:    task.StatusRunning,
		CreatedAt: time.Now(),
		Command:   opts.Command,
		Name:      opts.Name,
		Image:     opts.Image,
		Inputs:    make(map[task.Parameter]task.File, len(opts.Inputs)),
		Outputs:   make(map[task.Parameter]task.File, len(opts.Outputs)),
	}
	for name, path := range opts.Inputs {
		t.Inputs[task.Parameter(name)] = task.File{
			Path:     path,
			Provider: task.ProviderFS,
		}
	}
	for name, path := range opts.Outputs {
		t.Outputs[task.Parameter(name)] = task.File{
			Path:     path,
			Provider: task.ProviderFS,
		}
	}
	return t
}
