package service

import (
	"context"
	"errors"
	"strings"

	"github.com/sashank33v/recallengine/internal/model"
	"github.com/sashank33v/recallengine/internal/repository"
)

var ErrInvalidNote = errors.New("title is required (max 200 characters) and content is required (max 50000 characters)")

type NoteService struct {
	noteRepo *repository.NoteRepository
}

func NewNoteService(noteRepo *repository.NoteRepository) *NoteService {
	return &NoteService{noteRepo: noteRepo}
}

func (s *NoteService) Create(ctx context.Context, userID int64, title, content string) (*model.Note, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)

	if title == "" || len(title) > 200 || content == "" || len(content) > 50000 {
		return nil, ErrInvalidNote
	}
	return s.noteRepo.Create(ctx, userID, title, content)
}

func (s *NoteService) List(ctx context.Context, userID int64) ([]model.Note, error) {
	return s.noteRepo.ListByUser(ctx, userID)
}
