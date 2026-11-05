package apperror

import "errors"

var (
	ErrNoteNotFound         = errors.New("note not found")
	ErrJobInProgress        = errors.New("generation is already in progress for this note")
	ErrDraftNotFound        = errors.New("draft question not found")
	ErrConversationNotFound = errors.New("conversation not found")
	ErrConversationFull     = errors.New("this conversation is full, please start a new one")
)
