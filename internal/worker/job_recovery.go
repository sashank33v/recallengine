package worker

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type JobRecovery struct {
	db *sql.DB
}

func NewJobRecovery(db *sql.DB) *JobRecovery {
	return &JobRecovery{db: db}
}

func (r *JobRecovery) Run(ctx context.Context) {
	r.recover(ctx)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.recover(ctx)
		}
	}
}

func (r *JobRecovery) recover(ctx context.Context) {
	query := `UPDATE generation_jobs	SET status = 'failed', error = 'the worker stopped while generating, please try again', updated_at = now() WHERE status = 'running' AND updated_at < now() - interval '5 minutes'`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		log.Println("job recovery error:", err)
		return
	}

	recovered, err := result.RowsAffected()
	if err == nil && recovered > 0 {
		log.Println("marks stuck jobs as failed:", recovered)
	}
}
