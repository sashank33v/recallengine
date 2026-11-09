package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestConcurrentAnswerAreNotLost(t *testing.T) {
	url := os.Getenv("TEST_DATASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set, skipping database test")
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()
	db.SetMaxOpenConns(20)

	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		t.Fatalf("cannot reach the test database: %v", err)
	}
	users := NewUserRepository(db)
	questions := NewQuestionRepository(db)

	email := fmt.Sprintf("concurrency-test-%d@example.com", time.Now().UnixNano())
	user, err := users.Create(ctx, email, "not-a-real-hash")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		query1 := `DELETE FROM outbox_events WHERE (payload->>'user_id')::bigint = $1`
		query2 := `DELETE FROM users WHERE id = $1`
		db.ExecContext(ctx, query1, user.ID)
		db.ExecContext(ctx, query2, user.ID)
	}()
	question, err := questions.Create(ctx, user.ID, "General", "Test question?", "Test answer.")
	if err != nil {
		t.Fatal(err)
	}
	const workers = 10
	var wg sync.WaitGroup
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := questions.RecordAnswer(ctx, user.ID, question.ID, 5, time.Now())
			if err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("RecordAnswer failed: %v", err)
	}
	var repetitions int

	query3 := `SELECT repetitions FROM review_state WHERE question_id = $1`
	err = db.QueryRowContext(ctx, query3, question.ID).Scan(&repetitions)
	if err != nil {
		t.Fatal(err)
	}
	if repetitions != workers {
		t.Errorf("repetitions: got %d, want %d (an answer was lost)", repetitions, workers)
	}
	var answers int
	query4 := `SELECT COUNT(*) FROM answers WHERE question_id = $1`
	err = db.QueryRowContext(ctx, query4, question.ID).Scan(&answers)
	if err != nil {
		t.Fatal(err)
	}
	if answers != workers {
		t.Errorf("answers logged: got %d, want %d", answers, workers)
	}

	var events int
	query5 := `SELECT COUNT(*) FROM outbox_events WHERE event_type = 'AnswerSubmitted' AND (payload->>'user_id')::bigint = $1`
	err = db.QueryRowContext(ctx, query5, user.ID).Scan(&events)
	if err != nil {
		t.Fatal(err)
	}
	if events != workers {
		t.Errorf("outbox events: got %d, want %d", events, workers)
	}
}
