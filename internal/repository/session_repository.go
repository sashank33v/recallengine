package repository

import (
	"context"
	"database/sql"
	"time"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, tokenHash string, userID int64, expiresAt time.Time) error {
	query := `INSERT INTO sessions (token_hash, user_id, expires_at) VALUES( $1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, tokenHash, userID, expiresAt)
	return err
}

func (r *SessionRepository) GetUserID(ctx context.Context, tokenHash string) (int64, error) {
	query := `SELECT user_id FROM sessions WHERE token_hash = $1 AND expires_at > now()`

	var userID int64
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(&userID)
	return userID, err
}

func (r *SessionRepository) Delete(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM sessions WHERE token_hash = $1`
	_, err := r.db.ExecContext(ctx, query, tokenHash)
	return err
}
