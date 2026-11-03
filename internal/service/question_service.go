package service

import (
	"context"
	"errors"
	"strings"

	"github.com/sashank33v/recallengine/internal/model"
	"github.com/sashank33v/recallengine/internal/repository"
)

var ErrInvalidQuestion = errors.New("question is required (max 1000 characters) and answer is required (max 5000 characters)")
var ErrInvalidStatus = errors.New("status must be draft, approved or rejected")

type QuestionService struct {
	questionRepo *repository.QuestionRepository
}

func NewQuestionService(questionRepo *repository.QuestionRepository) *QuestionService {
	return &QuestionService{questionRepo: questionRepo}
}

func (s *QuestionService) Create(ctx context.Context, userID int64, question, answer string) (*model.Question, error) {
	question = strings.TrimSpace(question)
	answer = strings.TrimSpace(answer)

	if question == "" || len(question) > 1000 || answer == "" || len(answer) > 5000 {
		return nil, ErrInvalidQuestion
	}
	return s.questionRepo.Create(ctx, userID, question, answer)
}
func (s *QuestionService) List(ctx context.Context, userID int64, status string, limit, offset int) ([]model.Question, error) {
	if status != "" && status != "draft" && status != "approved" && status != "rejected" {
		return nil, ErrInvalidStatus
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.questionRepo.ListByUser(ctx, userID, status, limit, offset)
}
