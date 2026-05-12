package ipc

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

const runPath = "/run"

type Runner interface {
	Run(ctx context.Context, task *task.Task) error
}

type runHandler struct {
	runner Runner
	logger *slog.Logger
}

func newRunHandler(runner Runner, logger *slog.Logger) *runHandler {
	return &runHandler{
		runner: runner,
		logger: logger,
	}
}

type RunRequest struct {
	Command string            `json:"command"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Inputs  map[string]string `json:"inputs"`
	Outputs map[string]string `json:"outputs"`
}

type RunResponse struct {
	ID uuid.UUID `json:"id"`
}

func (h *runHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	t := taskFromRunRequest(req)

	if err := h.runner.Run(r.Context(), t); err != nil {
		h.logger.Error("failed to handle request", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := RunResponse{ID: t.ID}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func taskFromRunRequest(r RunRequest) *task.Task {
	return &task.Task{
		ID:        uuid.New(),
		Status:    task.StatusRunning,
		CreatedAt: time.Now(),
		Command:   strings.Replace(r.Command, "@", "./", -1),
		Name:      r.Name,
		Image:     r.Image,
		Inputs:    filesFromPaths(r.Inputs),
		Outputs:   filesFromPaths(r.Outputs),
	}
}

func filesFromPaths(paths map[string]string) map[string]task.File {
	files := make(map[string]task.File, len(paths))
	for name, path := range paths {
		files[name] = task.File{
			Path:     path,
			Provider: task.ProviderFS,
		}
	}
	return files
}
