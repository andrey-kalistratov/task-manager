package kafka

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

type Consumer struct {
	loop *Loop
}

func NewConsumer(cfg *config.MessageConfig, recorder Recorder, logger *slog.Logger) *Consumer {
	l := &Loop{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: cfg.Brokers,
			GroupID: cfg.GroupIDs.Planner,
			Topic:   cfg.Topics.Results,
		}),
		handler: &ResultHandler{
			recorder: recorder,
			logger:   logger.With("handler", "result"),
		},
		logger: logger,
	}
	return &Consumer{loop: l}
}

func (c *Consumer) Run(ctx context.Context) error {
	return c.loop.Run(ctx)
}

type Recorder interface {
	Record(ctx context.Context, result *task.Result) error
}

var _ Handler = (*ResultHandler)(nil)

type ResultHandler struct {
	recorder Recorder
	logger   *slog.Logger
}

func (h *ResultHandler) Handle(ctx context.Context, msg kafka.Message) {
	var dto Result
	if err := json.Unmarshal(msg.Value, &dto); err != nil {
		h.logger.Error("failed to unmarshal task result", "error", err)
		return
	}

	r, err := dto.toModel()
	if err != nil {
		h.logger.Error("failed to parse task result", "error", err)
		return
	}

	if err = h.recorder.Record(ctx, r); err != nil {
		h.logger.Error("failed to record task result", "error", err)
		return
	}
	h.logger.Info("task result recorded", "id", r.TaskID, "status", r.Status)
}

type Result struct {
	TaskID uuid.UUID `json:"task_id"`
	Status Status    `json:"status"`
}

func (r Result) toModel() (*task.Result, error) {
	status, err := r.Status.toModel()
	if err != nil {
		return nil, fmt.Errorf("parse result status: %w", err)
	}

	return &task.Result{
		TaskID: r.TaskID,
		Status: status,
	}, nil
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
