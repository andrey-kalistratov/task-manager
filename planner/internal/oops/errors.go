package oops

import (
	"fmt"

	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

type FileErrorKind int

const (
	FileErrorNotFound FileErrorKind = iota
	FileErrorPermission
)

type FileError struct {
	Kind FileErrorKind
	File task.File
}

func (e FileError) Error() string {
	switch e.Kind {
	case FileErrorNotFound:
		return fmt.Sprintf("file not found: %s", e.File.Path)
	case FileErrorPermission:
		return fmt.Sprintf("permission denied: %s", e.File.Path)
	default:
		return fmt.Sprintf("unknown file error: %s", e.File.Path)
	}
}
