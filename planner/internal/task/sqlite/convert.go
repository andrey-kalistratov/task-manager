package sqlite

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

type Task struct {
	ID        string
	Status    Status
	CreatedAt time.Time
	Command   string
	Name      string
	Image     string
	Inputs    []byte
	Uploads   []byte
	Downloads []byte
	Outputs   []byte
}

func newTask(t *task.Task) (*Task, error) {
	status, err := newStatus(t.Status)
	if err != nil {
		return nil, fmt.Errorf("serialize status: %w", err)
	}

	inputs, err := marshalParameters(t.Inputs)
	if err != nil {
		return nil, fmt.Errorf("marshal inputs: %w", err)
	}

	uploads, err := marshalParameters(t.Uploads)
	if err != nil {
		return nil, fmt.Errorf("marshal uploads: %w", err)
	}

	downloads, err := marshalParameters(t.Downloads)
	if err != nil {
		return nil, fmt.Errorf("marshal downloads: %w", err)
	}

	outputs, err := marshalParameters(t.Outputs)
	if err != nil {
		return nil, fmt.Errorf("marshal outputs: %w", err)
	}

	return &Task{
		ID:        t.ID.String(),
		Status:    status,
		CreatedAt: t.CreatedAt,
		Command:   t.Command,
		Name:      t.Name,
		Image:     t.Image,
		Inputs:    inputs,
		Uploads:   uploads,
		Downloads: downloads,
		Outputs:   outputs,
	}, nil
}

func marshalParameters(params map[task.Parameter]task.File) ([]byte, error) {
	raw := make(map[string]File)
	for p, f := range params {
		dto, err := newFile(f)
		if err != nil {
			return nil, fmt.Errorf("serialize file: %w", err)
		}

		raw[string(p)] = dto
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshal parameters: %w", err)
	}

	return data, nil
}

func (t *Task) toModel() (*task.Task, error) {
	id, err := uuid.Parse(t.ID)
	if err != nil {
		return nil, fmt.Errorf("parse task id: %w", err)
	}

	status, err := t.Status.toModel()
	if err != nil {
		return nil, fmt.Errorf("deserialize status: %w", err)
	}

	inputs, err := unmarshalParameters(t.Inputs)
	if err != nil {
		return nil, fmt.Errorf("unmarshal inputs: %w", err)
	}

	uploads, err := unmarshalParameters(t.Uploads)
	if err != nil {
		return nil, fmt.Errorf("unmarshal uploads: %w", err)
	}

	downloads, err := unmarshalParameters(t.Downloads)
	if err != nil {
		return nil, fmt.Errorf("unmarshal downloads: %w", err)
	}

	outputs, err := unmarshalParameters(t.Outputs)
	if err != nil {
		return nil, fmt.Errorf("unmarshal outputs: %w", err)
	}

	return &task.Task{
		ID:        id,
		Status:    status,
		CreatedAt: t.CreatedAt,
		Command:   t.Command,
		Name:      t.Name,
		Image:     t.Image,
		Inputs:    inputs,
		Uploads:   uploads,
		Downloads: downloads,
		Outputs:   outputs,
	}, nil
}

func unmarshalParameters(data []byte) (map[task.Parameter]task.File, error) {
	raw := make(map[string]File)
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal parameters: %w", err)
	}

	params := make(map[task.Parameter]task.File)
	for p, f := range raw {
		dto, err := f.toModel()
		if err != nil {
			return nil, fmt.Errorf("deserialize file: %w", err)
		}

		params[task.Parameter(p)] = dto
	}
	return params, nil
}

type Status string

const (
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

func newStatus(s task.Status) (Status, error) {
	switch s {
	case task.StatusRunning:
		return StatusRunning, nil
	case task.StatusSucceeded:
		return StatusSucceeded, nil
	case task.StatusFailed:
		return StatusFailed, nil
	default:
		return "", fmt.Errorf("unknown status: %v", s)
	}
}

func (s Status) toModel() (task.Status, error) {
	switch s {
	case StatusRunning:
		return task.StatusRunning, nil
	case StatusSucceeded:
		return task.StatusSucceeded, nil
	case StatusFailed:
		return task.StatusFailed, nil
	default:
		return 0, fmt.Errorf("unknown status: %v", s)
	}
}

type File struct {
	Path     string          `json:"path"`
	Provider StorageProvider `json:"provider"`
}

func newFile(f task.File) (File, error) {
	provider, err := newProvider(f.Provider)
	if err != nil {
		return File{}, fmt.Errorf("serialize provider: %w", err)
	}

	return File{
		Path:     f.Path,
		Provider: provider,
	}, nil
}

func (f File) toModel() (task.File, error) {
	provider, err := f.Provider.toModel()
	if err != nil {
		return task.File{}, fmt.Errorf("deserialize provider: %w", err)
	}

	return task.File{
		Path:     f.Path,
		Provider: provider,
	}, nil
}

type StorageProvider string

const (
	ProviderFS StorageProvider = "fs"
	ProviderS3 StorageProvider = "s3"
)

func newProvider(p task.StorageProvider) (StorageProvider, error) {
	switch p {
	case task.ProviderFS:
		return ProviderFS, nil
	case task.ProviderS3:
		return ProviderS3, nil
	default:
		return "", fmt.Errorf("unknown storage provider: %v", p)
	}
}

func (p StorageProvider) toModel() (task.StorageProvider, error) {
	switch p {
	case ProviderFS:
		return task.ProviderFS, nil
	case ProviderS3:
		return task.ProviderS3, nil
	default:
		return 0, fmt.Errorf("unknown storage provider: %v", p)
	}
}
