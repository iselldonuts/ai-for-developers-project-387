package errs

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrNotImplemented = errors.New("not implemented")
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
