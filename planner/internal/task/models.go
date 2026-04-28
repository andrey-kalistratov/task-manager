package task

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID        uuid.UUID
	Status    Status
	CreatedAt time.Time

	Command string
	Name    string
	Image   string

	Inputs    map[Parameter]File
	Uploads   map[Parameter]File
	Downloads map[Parameter]File
	Outputs   map[Parameter]File
}

type Status int

const (
	StatusRunning Status = iota
	StatusSucceeded
	StatusFailed
)

type Parameter string

type File struct {
	Path     string
	Provider StorageProvider
}

type StorageProvider int

const (
	ProviderFS StorageProvider = iota
	ProviderS3
)

type Result struct {
	TaskID uuid.UUID
	Status Status
}
