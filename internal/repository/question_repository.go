package repository

import (
	"context"
	"database/sql"

	"github.com/sashank33v/recallengine/internal/model"
)

type QuestionRepository struct {
	db *sql.DB
}

func NewQuestionRepository(db *sql.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

func (r *QuestionRepository) Create(ctx context.Context, userID int64, questionText, answerText string) (*model.Question, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO questions (user_id, question_text, answer_text)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, question_text, answer_text, source, status, created_at`

	q := &model.Question{}
	err = tx.QueryRowContext(ctx, query, userID, questionText, answerText).Scan(
		&q.ID, &q.UserID, &q.QuestionText, &q.AnswerText, &q.Source, &q.Status, &q.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO review_state (question_id, user_id) VALUES ($1, $2)`,
		q.ID, userID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return q, nil
}
