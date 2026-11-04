package repository

import (
	"context"
	"database/sql"

	"github.com/sashank33v/recallengine/internal/model"
)

type NoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) Create(ctx context.Context, userID int64, title, content string, chunks []string) (*model.Note, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	query := `
		INSERT INTO notes (user_id, title, content)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, title, content, created_at, updated_at`

	note := &model.Note{}
	err = r.db.QueryRowContext(ctx, query, userID, title, content).Scan(
		&note.ID, &note.UserID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	for i, chunk := range chunks {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO note_chunks (note_id, user_id, chunk_index, content) VALUES ($1, $2, $3, $4)`,
			note.ID, userID, i, chunk)
		if err != nil {
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return note, nil
}

func (r *NoteRepository) ListByUser(ctx context.Context, userID int64) ([]model.Note, error) {
	query := `
		SELECT id, user_id, title, content, created_at, updated_at
		FROM notes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []model.Note{}
	for rows.Next() {
		var n model.Note
		err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}
