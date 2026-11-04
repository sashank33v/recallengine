package repository

import (
	"context"
	"database/sql"

	"github.com/sashank33v/recallengine/internal/model"
)

type MasteryRepository struct {
	db *sql.DB
}

func NewMasteryRepository(db *sql.DB) *MasteryRepository {
	return &MasteryRepository{db: db}
}

func (r *MasteryRepository) ListByUser(ctx context.Context, userID int64) ([]model.TopicMastery, error) {
	query := `SELECT t.name, m.total_answers, m.correct_answers FROM topic_mastery m JOIN topics t ON t.id = m.topic_id	WHERE m.user_id = $1 ORDER BY m.correct_answers::float / m.total_answers ASC, t.name ASC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []model.TopicMastery{}

	for rows.Next() {
		var m model.TopicMastery
		err := rows.Scan(&m.Topic, &m.TotalAnswers, &m.CorrectAnswers)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}
