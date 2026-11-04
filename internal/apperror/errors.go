package apperror

import "errors"

var (
	ErrNoteNotFound  = errors.New("note not found")
	ErrJobInProgress = errors.New("generation is already in progress for this note")
)
