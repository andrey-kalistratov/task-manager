package task

type FileErrorKind int

const (
	FileErrorNotFound FileErrorKind = iota
	FileErrorPermission
)

type FileError struct {
	Kind FileErrorKind
	File File
	Err  error
}

func (e *FileError) Error() string {
	return e.Err.Error()
}
