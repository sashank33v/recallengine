package service

import (
	"context"

	"github.com/sashank33v/recallengine/internal/apperror"
	"github.com/sashank33v/recallengine/internal/model"
	"github.com/sashank33v/recallengine/internal/repository"
)

const maxConversationMessages = 40

type ConversationService struct {
	repo *repository.ConversationRepository
}

func NewConversationService(repo *repository.ConversationRepository) *ConversationService {
	return &ConversationService{repo: repo}
}

func (s *ConversationService) Start(ctx context.Context, userID, conversationID int64, question string) (int64, error) {
	if conversationID == 0 {
		return s.repo.Create(ctx, userID, truncateRunes(question, 60))
	}

	owns, err := s.repo.Owns(ctx, userID, conversationID)
	if err != nil {
		return 0, err
	}
	if !owns {
		return 0, apperror.ErrConversationNotFound
	}
	count, err := s.repo.CountMessages(ctx, conversationID)
	if err != nil {
		return 0, err
	}
	if count >= maxConversationMessages {
		return 0, apperror.ErrConversationFull
	}
	return conversationID, nil
}

func (s *ConversationService) AddMessage(ctx context.Context, conversationID int64, role, content string) error {
	return s.repo.AddMessage(ctx, conversationID, role, content)
}

func (s *ConversationService) List(ctx context.Context, userID int64) ([]model.Conversation, error) {
	return s.repo.List(ctx, userID)
}

func (s *ConversationService) Messages(ctx context.Context, userID, conversationID int64) ([]model.ChatMessage, error) {
	owns, err := s.repo.Owns(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, apperror.ErrConversationNotFound
	}
	return s.repo.Messages(ctx, conversationID)
}

func truncateRunes(text string, max int) string {
	runes := []rune(text)
	if len(runes) > max {
		return string(runes[:max])
	}
	return text
}
