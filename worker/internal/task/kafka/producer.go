package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/andrey-kalistratov/task-manager/worker/internal/config"
	"github.com/andrey-kalistratov/task-manager/worker/internal/task"
)

var _ task.Publisher = (*Producer)(nil)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(cfg *config.MessageConfig) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:  kafka.TCP(cfg.Brokers...),
			Topic: cfg.Topics.Results,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, r task.Result) error {
	dto, err := NewResult(r)
	if err != nil {
		return fmt.Errorf("encode result: %w", err)
	}

	value, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(r.TaskID.String()),
		Value: value,
	}
	if err = p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("write to kafka: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

type Result struct {
	TaskID uuid.UUID `json:"task_id"`
	Status Status    `json:"status"`
}

func NewResult(result task.Result) (Result, error) {
	status, err := NewStatus(result.Status)
	if err != nil {
		return Result{}, fmt.Errorf("encode status: %w", err)
	}

	return Result{
		TaskID: result.TaskID,
		Status: status,
	}, nil
}

type Status string

const (
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

func NewStatus(status task.Status) (Status, error) {
	switch status {
	case task.StatusSucceeded:
		return StatusSucceeded, nil
	case task.StatusFailed:
		return StatusFailed, nil
	default:
		return "", fmt.Errorf("unknown status: %v", status)
	}
}
