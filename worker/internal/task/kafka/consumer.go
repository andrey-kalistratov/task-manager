package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/andrey-kalistratov/task-manager/worker/internal/config"
	"github.com/andrey-kalistratov/task-manager/worker/internal/task"
)

type Consumer struct {
	loop *Loop
}

func NewConsumer(cfg *config.MessageConfig, processor Processor, logger *slog.Logger) *Consumer {
	l := &Loop{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: cfg.Brokers,
			GroupID: cfg.GroupIDs.Worker,
			Topic:   cfg.Topics.Tasks,
		}),
		handler: &TaskHandler{
			processor: processor,
			logger:    logger.With("handler", "task"),
		},
		logger: logger,
	}
	return &Consumer{loop: l}
}

func (c *Consumer) Run(ctx context.Context) error {
	return c.loop.Run(ctx)
}

type Processor interface {
	Process(ctx context.Context, task *task.Task) error
}

var _ Handler = (*TaskHandler)(nil)

type TaskHandler struct {
	processor Processor
	logger    *slog.Logger
}

func (h *TaskHandler) Handle(ctx context.Context, msg kafka.Message) {
	var dto Task
	if err := json.Unmarshal(msg.Value, &dto); err != nil {
		h.logger.Error("failed to unmarshal task", "error", err)
		return
	}

	t, err := dto.toModel()
	if err != nil {
		h.logger.Error("failed to parse task", "error", err)
	}

	if err = h.processor.Process(ctx, t); err != nil {
		h.logger.Error("failed to process task", "error", err)
		return
	}
	h.logger.Info("task processed", "id", t.ID)
}

type Task struct {
	ID      uuid.UUID       `json:"id"`
	Command string          `json:"command"`
	Inputs  map[string]File `json:"inputs"`
	Outputs map[string]File `json:"outputs"`
}

func (t *Task) toModel() (*task.Task, error) {
	inputs := make(map[string]task.File, len(t.Inputs))
	for path, f := range t.Inputs {
		file, err := f.toModel()
		if err != nil {
			return nil, fmt.Errorf("parse file: %w", err)
		}

		inputs[path] = *file
	}

	outputs := make(map[string]task.File, len(t.Outputs))
	for path, f := range t.Outputs {
		file, err := f.toModel()
		if err != nil {
			return nil, fmt.Errorf("parse file: %w", err)
		}

		outputs[path] = *file
	}

	return &task.Task{
		ID:      t.ID,
		Command: t.Command,
		Inputs:  inputs,
		Outputs: outputs,
	}, nil
}

type File struct {
	Path     string          `json:"path"`
	Provider StorageProvider `json:"provider"`
}

func (f *File) toModel() (*task.File, error) {
	provider, err := f.Provider.toModel()
	if err != nil {
		return nil, fmt.Errorf("parse storage provider: %w", err)
	}

	return &task.File{
		Path:     f.Path,
		Provider: provider,
	}, nil
}

type StorageProvider string

const (
	ProviderS3   StorageProvider = "s3"
	ProviderTask StorageProvider = "task"
)

func (s StorageProvider) toModel() (task.StorageProvider, error) {
	switch s {
	case ProviderS3:
		return task.ProviderS3, nil
	case ProviderTask:
		return task.ProviderTask, nil
	default:
		return 0, fmt.Errorf("unknown storage provider: %v", s)
	}
}
