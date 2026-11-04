package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/sashank33v/recallengine/internal/model"
	"github.com/sashank33v/recallengine/internal/scheduler"
)

type QuestionRepository struct {
	db *sql.DB
}

func NewQuestionRepository(db *sql.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

func (r *QuestionRepository) Create(ctx context.Context, userID int64, topicName, questionText, answerText string) (*model.Question, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var topicID int
	query1 := `INSERT INTO topics (user_id, name) VALUES ($1, $2) ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name RETURNING id`
	err = tx.QueryRowContext(ctx, query1, userID, topicName).Scan(&topicID)
	if err != nil {
		return nil, err
	}
	query2 := `INSERT INTO questions (user_id, topic_id, question_text, answer_text) VALUES ($1, $2, $3, $4) RETURNING id, user_id, question_text, answer_text, source, status, created_at`
	q := &model.Question{}
	err = tx.QueryRowContext(ctx, query2, userID, topicID, questionText, answerText).Scan(&q.ID, &q.UserID, &q.QuestionText, &q.AnswerText, &q.Source, &q.Status, &q.CreatedAt)
	if err != nil {
		return nil, err
	}
	query3 := `INSERT INTO review_state (question_id, user_id) VALUES ($1, $2)`
	_, err = tx.ExecContext(ctx, query3, q.ID, userID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *QuestionRepository) ListDue(ctx context.Context, userID int64, limit int) ([]model.Question, error) {
	query := `SELECT q.id, q.user_id, q.question_text, q.answer_text, q.source, q.status, q.created_at
		FROM questions q
		JOIN review_state r ON r.question_id = q.id
		WHERE q.user_id = $1 AND q.status = 'approved' AND r.due_at <= now()
		ORDER BY r.due_at ASC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := []model.Question{}

	for rows.Next() {
		var q model.Question
		err := rows.Scan(&q.ID, &q.UserID, &q.QuestionText, &q.AnswerText, &q.Source, &q.Status, &q.CreatedAt)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

func (r *QuestionRepository) RecordAnswer(ctx context.Context, userID, questionID int64, quality int, now time.Time) (scheduler.State, time.Time, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return scheduler.State{}, time.Time{}, err
	}
	defer tx.Rollback()
	query1 := `SELECT ease_factor, interval_days, repetitions FROM review_state WHERE question_id = $1 AND user_id = $2 FOR UPDATE`
	var state scheduler.State
	err = tx.QueryRowContext(ctx, query1, questionID, userID).Scan(&state.EaseFactor, &state.IntervalDays, &state.Repetitions)
	if err != nil {
		return scheduler.State{}, time.Time{}, err
	}

	newState, dueAt := scheduler.Next(state, quality, now)
	query2 := `UPDATE review_state SET ease_factor = $1, interval_days = $2, repetitions = $3, due_at = $4, last_reviewed_at = $5 WHERE question_id = $6`
	_, err = tx.ExecContext(ctx, query2, newState.EaseFactor, newState.IntervalDays, newState.Repetitions, dueAt, now, questionID)

	if err != nil {
		return scheduler.State{}, time.Time{}, err
	}
	query3 := `INSERT INTO answers (user_id, question_id, quality, answered_at) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, query3, userID, questionID, quality, now)
	if err != nil {
		return scheduler.State{}, time.Time{}, err
	}

	payload, err := json.Marshal(map[string]any{
		"user_id":     userID,
		"question_id": questionID,
		"quality":     quality,
		"answered_at": now,
	})
	if err != nil {
		return scheduler.State{}, time.Time{}, err
	}
	query4 := `INSERT INTO outbox_events (topic, event_key, event_type, payload) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, query4, "answer.events", strconv.FormatInt(userID, 10), "AnswerSubmitted", string(payload))
	if err != nil {
		return scheduler.State{}, time.Time{}, err
	}
	err = tx.Commit()
	if err != nil {
		return scheduler.State{}, time.Time{}, err
	}
	return newState, dueAt, nil

}
func (r *QuestionRepository) ListByUser(ctx context.Context, userID int64, status string, limit, offset int) ([]model.Question, error) {
	query := `
		SELECT id, user_id, question_text, answer_text, source, status, created_at
		FROM questions
		WHERE user_id = $1 AND ($2::text = '' OR status = $2::text)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, query, userID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := []model.Question{}
	for rows.Next() {
		var q model.Question
		err := rows.Scan(&q.ID, &q.UserID, &q.QuestionText, &q.AnswerText, &q.Source, &q.Status, &q.CreatedAt)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}
