package worker

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type SessionCleaner struct {
	db *sql.DB
}

func NewSessionCleaner(db *sql.DB) *SessionCleaner {
	return &SessionCleaner{db: db}
}

func (c *SessionCleaner) Run(ctx context.Context) {
	c.clean(ctx)

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.clean(ctx)
		}
	}
}

func (c *SessionCleaner) clean(ctx context.Context) {
	query := `DELETE FROM sessions WHERE expires_at < now()`
	result, err := c.db.ExecContext(ctx, query)
	if err != nil {
		log.Println("session cleanup error:", err)
		return
	}
	deleted, err := result.RowsAffected()
	if err == nil && deleted > 0 {
		log.Println("deleted expired sessions:", deleted)
	}
}
