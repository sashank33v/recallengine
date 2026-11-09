package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestDueQuestionsPutWeakTopicsFirst(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")

	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set, skipping database test")
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	err = db.PingContext(ctx)
	if err != nil {
		t.Fatalf("cannot reach the test database: %v", err)
	}
	users := NewUserRepository(db)
	questions := NewQuestionRepository(db)

	email := fmt.Sprintf("priority-test-%dexample.com", time.Now().UnixNano())
	user, err := users.Create(ctx, email, "not-a-real-hash")
	if err != nil {
		t.Fatal(err)
	}
	query := `DELETE FROM users WHERE id = $1`
	defer db.ExecContext(ctx, query, user.ID)

	_, err = questions.Create(ctx, user.ID, "Strong", "Strong question?", "A")
	if err != nil {
		t.Fatal(err)
	}

	_, err = questions.Create(ctx, user.ID, "Weak", "Weak question?", "A")
	if err != nil {
		t.Fatal(err)
	}

	setMastery := func(topic string, total, correct int) {
		query := `INSERT INTO topic_mastery (user_id, topic_id, total_answers, correct_answers) SELECT $1::bigint, id, $2::int, $3::int FROM topics WHERE user_id = $1 AND name = $4`
		_, err := db.ExecContext(ctx, query, user.ID, total, correct, topic)
		if err != nil {
			t.Fatal(err)
		}
	}
	setMastery("Strong", 4, 4)
	setMastery("Weak", 4, 1)

	due, err := questions.ListDue(ctx, user.ID, 10)

	if err != nil {
		t.Fatal(err)
	}

	if len(due) != 2 {
		t.Fatalf("due questions: got %d, want 2", len(due))
	}
	if due[0].QuestionText != "Weak question?" {
		t.Errorf("first question: got %q, want the weak topic first", due[0].QuestionText)
	}
}
