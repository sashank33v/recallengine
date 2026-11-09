package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/sashank33v/recallengine/internal/service"
)

type GenerationWorker struct {
	db        *sql.DB
	generator *service.QuestionGenerator
}

func NewGenerationWorker(db *sql.DB, generator *service.QuestionGenerator) *GenerationWorker {
	return &GenerationWorker{db: db, generator: generator}
}

type generationEvent struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	Payload   struct {
		JobID  int64 `json:"job_id"`
		UserID int64 `json:"user_id"`
		NoteID int64 `json:"note_id"`
	} `json:"payload"`
}

func (g *GenerationWorker) Handle(ctx context.Context, value []byte) error {
	var event generationEvent
	err := json.Unmarshal(value, &event)
	if err != nil {
		return err
	}
	if event.EventType != "GenerationRequested" {
		return nil
	}
	jobID := event.Payload.JobID
	query1 := `UPDATE generation_jobs SET status = 'running', updated_at = now() WHERE id = $1 AND status = 'pending'`
	result, err := g.db.ExecContext(ctx, query1, jobID)
	if err != nil {
		return err
	}
	claimed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if claimed == 0 {
		log.Println("job already handled, skipping:", jobID)
		return nil
	}

	count, err := g.generate(ctx, jobID, event.Payload.UserID, event.Payload.NoteID)
	if err != nil {
		if ctx.Err() != nil {

			g.setStatus(context.Background(), jobID, "pending", "")
			return ctx.Err()
		}
		log.Printf("job %d failed: %v", jobID, err)
		g.setStatus(ctx, jobID, "failed", err.Error())
		return nil
	}

	log.Printf("job %d done: %d draft questions", jobID, count)
	return nil
}

func (g *GenerationWorker) generate(ctx context.Context, jobID, userID, noteID int64) (int, error) {

	var title, content string
	query1 := `SELECT title, content FROM notes WHERE id = $1 AND user_id = $2`
	err := g.db.QueryRowContext(ctx, query1, noteID, userID).Scan(&title, &content)
	if err != nil {
		return 0, err
	}

	aiCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	questions, err := g.generator.Generate(aiCtx, title, content)
	if err != nil {
		return 0, err
	}

	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	topicName := []rune(title)
	if len(topicName) > 100 {
		topicName = topicName[:100]
	}

	var topicID int64
	query2 := `INSERT INTO topics (user_id, name) VALUES ($1, $2) ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name RETURNING id`
	err = tx.QueryRowContext(ctx, query2, userID, string(topicName)).Scan(&topicID)
	if err != nil {
		return 0, err
	}
	created := 0

	for _, q := range questions {
		var exists bool
		query3 := `SELECT EXISTS (SELECT 1 FROM questions WHERE user_id = $1 AND lower(btrim(regexp_replace(question_text, '[^[:alnum:]]+', ' ', 'g'))) = lower(btrim(regexp_replace($2::text, '[^[:alnum:]]+', ' ', 'g'))))`
		err = tx.QueryRowContext(ctx, query3, userID, q.Question).Scan(&exists)
		if err != nil {
			return 0, err
		}
		if exists {
			continue
		}
		query4 := `INSERT INTO questions (user_id, note_id, topic_id, question_text, answer_text, source, status) VALUES ($1, $2, $3, $4, $5, 'ai', 'draft')`
		_, err := tx.ExecContext(ctx, query4, userID, noteID, topicID, q.Question, q.Answer)
		if err != nil {
			return 0, err
		}
		created++
	}

	query4 := `UPDATE generation_jobs SET status = 'done', questions_created = $1, updated_at = now() WHERE id = $2`

	_, err = tx.ExecContext(ctx, query4, len(questions), jobID)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return created, nil
}

func (g *GenerationWorker) setStatus(ctx context.Context, jobID int64, status, errMsg string) {
	text := []rune(errMsg)
	if len(text) > 500 {
		errMsg = string(text[:500])
	}
	query := `UPDATE generation_jobs SET status = $1, error = NULLIF($2, ''), updated_at = now() WHERE id = $3`

	_, err := g.db.ExecContext(ctx, query, status, errMsg, jobID)
	if err != nil {
		log.Println("could not update job status:", err)
	}
}
