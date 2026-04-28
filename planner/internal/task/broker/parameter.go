package broker

import (
	"fmt"

	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

type Parameter struct {
	Name   string `json:"param"`
	Source File   `json:"source"`
}

func newParameter(p task.Parameter, source task.File) (Parameter, error) {
	file, err := newFile(source)
	if err != nil {
		return Parameter{}, fmt.Errorf("serialize file %s: %w", source.Path, err)
	}

	return Parameter{
		Name:   string(p),
		Source: file,
	}, nil
}

func (p *Parameter) toModel() (task.Parameter, task.File, error) {
	file, err := p.Source.toModel()
	if err != nil {
		return "", task.File{}, fmt.Errorf("deserialize file %s: %w", p.Source.Path, err)
	}
	return task.Parameter(p.Name), file, nil
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

func (f *File) toModel() (task.File, error) {
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
