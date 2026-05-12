package kafka

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

func NewProducer(cfg *config.MessageConfig) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:  kafka.TCP(cfg.Brokers...),
			Topic: cfg.Topics.Tasks,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, t *task.Task) error {
	dto, err := newTask(t)
	if err != nil {
		return fmt.Errorf("encode task: %w", err)
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
	ID      uuid.UUID       `json:"id"`
	Command string          `json:"command"`
	Inputs  map[string]File `json:"inputs"`
	Outputs map[string]File `json:"outputs"`
}

func newTask(t *task.Task) (*Task, error) {
	uploads, err := buildFiles(t.Uploads)
	if err != nil {
		return nil, fmt.Errorf("encode uploads: %w", err)
	}

	downloads, err := buildFiles(t.Downloads)
	if err != nil {
		return nil, fmt.Errorf("encode downloads: %w", err)
	}

	return &Task{
		ID:      t.ID,
		Command: t.Command,
		Inputs:  uploads,
		Outputs: downloads,
	}, nil
}

func buildFiles(files map[string]task.File) (map[string]File, error) {
	result := make(map[string]File)
	for path, file := range files {
		f, err := newFile(file)
		if err != nil {
			return nil, fmt.Errorf("encode file: %w", err)
		}

		result[path] = *f
	}
	return result, nil
}

type File struct {
	Path     string          `json:"path"`
	Provider StorageProvider `json:"provider"`
}

func newFile(f task.File) (*File, error) {
	provider, err := newStorageProvider(f.Provider)
	if err != nil {
		return nil, fmt.Errorf("encode storage provider: %w", err)
	}

	return &File{
		Path:     f.Path,
		Provider: provider,
	}, nil
}

type StorageProvider string

const (
	ProviderFS StorageProvider = "fs"
	ProviderS3 StorageProvider = "s3"
)

func newStorageProvider(p task.StorageProvider) (StorageProvider, error) {
	switch p {
	case task.ProviderFS:
		return ProviderFS, nil
	case task.ProviderS3:
		return ProviderS3, nil
	default:
		return "", fmt.Errorf("unknown storage provider: %v", p)
	}
}
