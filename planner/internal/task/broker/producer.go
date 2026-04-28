package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

var _ task.Publisher = (*Producer)(nil)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(cfg *config.Config) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:  kafka.TCP(cfg.Messaging.Brokers...),
			Topic: cfg.Messaging.Topics.Tasks,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, t *task.Task) error {
	dto, err := newTask(t)
	if err != nil {
		return fmt.Errorf("serialize task: %w", err)
	}

	value, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(t.ID.String()),
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

type Task struct {
	ID      uuid.UUID   `json:"id"`
	Command string      `json:"command"`
	Fetch   []Parameter `json:"fetch"`
	Upload  []string    `json:"upload"`
}

func newTask(t *task.Task) (*Task, error) {
	dto := &Task{
		ID:      t.ID,
		Command: t.Command,
		Fetch:   make([]Parameter, 0, len(t.Uploads)),
		Upload:  make([]string, 0, len(t.Downloads)),
	}
	for name, file := range t.Uploads {
		param, err := newParameter(name, file)
		if err != nil {
			return nil, fmt.Errorf("serialize input parameter: %w", err)
		}

		dto.Fetch = append(dto.Fetch, param)
	}
	for name := range t.Downloads {
		dto.Upload = append(dto.Upload, string(name))
	}
	return dto, nil
}
