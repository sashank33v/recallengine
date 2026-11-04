package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/sashank33v/recallengine/internal/apperror"
)

type GenerationRepository struct {
	db *sql.DB
}

func NewGenerationRepository(db *sql.DB) *GenerationRepository {
	return &GenerationRepository{db: db}
}

func (r *GenerationRepository) CreateJob(ctx context.Context, userID, noteID int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var noteExists bool
	query1 := `SELECT EXISTS (SELECT 1 FROM notes WHERE id = $1 AND user_id = $2)`
	err = tx.QueryRowContext(ctx, query1, noteID, userID).Scan(&noteExists)

	if err != nil {
		return 0, err
	}

	if !noteExists {
		return 0, apperror.ErrNoteNotFound
	}

	var inProgress bool
	query2 := `SELECT EXISTS (SELECT 1 FROM generation_jobs WHERE note_id = $1 AND status IN ('pending', 'running'))`
	err = tx.QueryRowContext(ctx, query2, noteID).Scan(&inProgress)

	if err != nil {
		return 0, err
	}

	if inProgress {
		return 0, apperror.ErrJobInProgress
	}

	var jobID int64
	query3 := `INSERT INTO generation_jobs (user_id, note_id) VALUES ($1, $2) RETURNING id`
	err = tx.QueryRowContext(ctx, query3, userID, noteID).Scan(&jobID)
	if err != nil {
		return 0, err
	}

	payload, err := json.Marshal(map[string]any{
		"job_id":  jobID,
		"user_id": userID,
		"note_id": noteID,
	})

	if err != nil {
		return 0, err
	}

	query4 := `INSERT INTO outbox_events (topic, event_key, event_type, payload) VALUES ($1, $2, $3, $4)`

	_, err = tx.ExecContext(ctx, query4, "generation.jobs", strconv.FormatInt(userID, 10), "GenerationRequested", string(payload))

	if err != nil {
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return jobID, nil
}
