package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/sashank33v/recallengine/internal/model"
	"github.com/sashank33v/recallengine/internal/repository"
	"github.com/sashank33v/recallengine/internal/scheduler"
)

var (
	ErrInvalidQuality   = errors.New("Quality must be between 0 and 5")
	ErrQuestionNotFound = errors.New("Question not found")
)

type StudyService struct {
	questionRepo *repository.QuestionRepository
}

func NewStudyService(questionRepo *repository.QuestionRepository) *StudyService {
	return &StudyService{questionRepo: questionRepo}
}

func (s *StudyService) DueQuestions(ctx context.Context, userID int64) ([]model.Question, error) {
	return s.questionRepo.ListDue(ctx, userID, 20)
}

func (s *StudyService) Answer(ctx context.Context, userID, questionID int64, quality int) (scheduler.State, time.Time, error) {
	if quality < 0 || quality > 5 {
		return scheduler.State{}, time.Time{}, ErrInvalidQuality
	}

	state, dueAt, err := s.questionRepo.RecordAnswer(ctx, userID, questionID, quality, time.Now())
	if errors.Is(err, sql.ErrNoRows) {
		return scheduler.State{}, time.Time{}, ErrQuestionNotFound
	}
	return state, dueAt, err
}
