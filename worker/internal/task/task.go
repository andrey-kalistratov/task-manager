package task

import "github.com/google/uuid"

type Task struct {
	ID      uuid.UUID
	Command string
	Inputs  map[string]File
	Outputs map[string]File
}

type File struct {
	Path     string
	Provider StorageProvider
}

type StorageProvider int

const (
	ProviderS3 StorageProvider = iota
	ProviderTask
)

type Result struct {
	TaskID uuid.UUID
	Status Status
}

type Status int

const (
	StatusSucceeded Status = iota
	StatusFailed
)
