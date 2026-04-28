package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

var _ task.Transport = (*ResultConsumer)(nil)

type ResultConsumer struct {
	receiver *Receiver
}

func NewResultConsumer(cfg *config.Config, service task.Service, logger *slog.Logger) *ResultConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Messaging.Brokers,
		GroupID: cfg.Messaging.GroupIDs.Planner,
		Topic:   cfg.Messaging.Topics.Results,
	})
	h := &resultHandler{
		service: service,
		logger:  logger.With("handler", "result"),
	}
	return &ResultConsumer{receiver: NewReceiver(r, h, logger)}
}

func (c *ResultConsumer) Serve(ctx context.Context) error {
	return c.receiver.Receive(ctx)
}

func (c *ResultConsumer) Shutdown(ctx context.Context) error {
	return c.receiver.Shutdown(ctx)
}

var _ Handler = (*resultHandler)(nil)

type resultHandler struct {
	service task.Service
	logger  *slog.Logger
}

func (h *resultHandler) Handle(ctx context.Context, msg kafka.Message) {
	var dto Result
	if err := json.Unmarshal(msg.Value, &dto); err != nil {
		h.logger.Error("failed to unmarshal task result", "error", err)
		return
	}

	r, err := dto.toModel()
	if err != nil {
		h.logger.Error("failed to deserialize task result", "error", err)
		return
	}

	if err = h.service.Record(ctx, r); err != nil {
		h.logger.Error("failed to record task result", "error", err)
	}
	h.logger.Info("task done", "id", r.TaskID, "status", r.Status)
}

type Result struct {
	TaskID uuid.UUID   `json:"task_id"`
	Status Status      `json:"status"`
	Fetch  []Parameter `json:"fetch"`
}

func (r Result) toModel() (task.Result, error) {
	result := task.Result{
		TaskID:    r.TaskID,
		Downloads: make(map[task.Parameter]task.File),
	}

	status, err := r.Status.toModel()
	if err != nil {
		return task.Result{}, fmt.Errorf("deserialize result status: %w", err)
	}
	result.Status = status

	for _, param := range r.Fetch {
		name, file, err := param.toModel()
		if err != nil {
			return task.Result{}, fmt.Errorf("deserialize output parameter: %w", err)
		}
		result.Downloads[name] = file
	}
	return result, nil
}

type Status string

const (
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

func (s Status) toModel() (task.Status, error) {
	switch s {
	case StatusSucceeded:
		return task.StatusSucceeded, nil
	case StatusFailed:
		return task.StatusFailed, nil
	default:
		return 0, fmt.Errorf("unknown status: %v", s)
	}
}
