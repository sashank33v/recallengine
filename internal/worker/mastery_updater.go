package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

type MasteryUpdater struct {
	db *sql.DB
}
type answerPayload struct {
	UserID     int64     `json:"user_id"`
	QuestionID int64     `json:"question_id"`
	Quality    int       `json:"quality"`
	AnsweredAt time.Time `json:"answered_at"`
}
type answerEvent struct {
	EventID   string        `json:"event_id"`
	EventType string        `json:"event_type"`
	Payload   answerPayload `json:"payload"`
}

func NewMasteryUpdater(db *sql.DB) *MasteryUpdater {
	return &MasteryUpdater{db: db}
}

func (m *MasteryUpdater) Handle(ctx context.Context, value []byte) error {
	var event answerEvent

	err := json.Unmarshal(value, &event)
	if err != nil {
		log.Println("skipping unreadable message:", err)
		return nil
	}

	if event.EventType != "AnswerSubmitted" {
		return nil
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `INSERT INTO processed_events (event_id, consumer) VALUES ($1, $2) ON CONFLICT DO NOTHING`, event.EventID, "mastery")
	if err != nil {
		return err
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if inserted == 0 {
		log.Println("duplicate event, skipping:", event.EventID)
		return nil
	}

	correct := 0
	if event.Payload.Quality >= 3 {
		correct = 1
	}
	query1 := `INSERT INTO topic_mastery (user_id, topic_id, total_answers, correct_answers, last_answered_at) SELECT $1::bigint, q.topic_id, 1, $2::int, $3::timestamptz FROM questions q WHERE q.id = $4 AND q.user_id = $1 AND q.topic_id IS NOT NULL ON CONFLICT (user_id, topic_id) DO UPDATE SET total_answers    = topic_mastery.total_answers + 1, correct_answers  = topic_mastery.correct_answers + EXCLUDED.correct_answers, last_answered_at = EXCLUDED.last_answered_at`
	_, err = tx.ExecContext(ctx, query1, event.Payload.UserID, correct, event.Payload.AnsweredAt, event.Payload.QuestionID)
	if err != nil {
		return err
	}

	return tx.Commit()

}
