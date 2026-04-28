package oops

import (
	"fmt"

	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

type FileErrorKind int

const (
	FileErrorUnknown FileErrorKind = iota
	FileErrorNotFound
	FileErrorPermission
)

type FileError struct {
	Kind FileErrorKind
	File task.File
	Err  error
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

type ParameterErrorKind int

const (
	ParameterErrorOutputNotFound ParameterErrorKind = iota
)

type ParameterError struct {
	Kind  ParameterErrorKind
	Param task.Parameter
}

func (e ParameterError) Error() string {
	switch e.Kind {
	case ParameterErrorOutputNotFound:
		return fmt.Sprintf("file of output parameter %q not found in results", e.Param)
	default:
		return fmt.Sprintf("unknown parameter error: %q", e.Param)
	}
}
