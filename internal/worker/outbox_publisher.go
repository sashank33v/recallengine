package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/sashank33v/recallengine/internal/platform/kafka"
)

type OutboxPublisher struct {
	db       *sql.DB
	producer *kafka.Producer
}

func NewOutboxPublisher(db *sql.DB, producer *kafka.Producer) *OutboxPublisher {
	return &OutboxPublisher{db: db, producer: producer}
}

type outboxEvent struct {
	id        int64
	eventID   string
	topic     string
	key       string
	eventType string
	payload   string
}

type envelope struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

func (p *OutboxPublisher) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := p.publishPending(ctx)
			if err != nil {
				log.Println("outbox publish error:", err)
			}
		}
	}
}

func (p *OutboxPublisher) publishPending(ctx context.Context) error {
	events, err := p.loadPending(ctx)
	if err != nil {
		return err
	}

	for _, e := range events {
		message, err := json.Marshal(envelope{
			EventID:   e.eventID,
			EventType: e.eventType,
			Payload:   json.RawMessage(e.payload),
		})
		if err != nil {
			return err
		}

		err = p.producer.Publish(ctx, e.topic, e.key, message)
		if err != nil {
			return err
		}

		_, err = p.db.ExecContext(ctx,
			`UPDATE outbox_events SET published_at = now() WHERE id = $1`, e.id)
		if err != nil {
			return err
		}
	}
	return nil
}
func (p *OutboxPublisher) loadPending(ctx context.Context) ([]outboxEvent, error) {
	query := `
		SELECT id, event_id::text, topic, event_key, event_type, payload::text
		FROM outbox_events
		WHERE published_at IS NULL
		ORDER BY id
		LIMIT 50`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []outboxEvent{}
	for rows.Next() {
		var e outboxEvent
		err := rows.Scan(&e.id, &e.eventID, &e.topic, &e.key, &e.eventType, &e.payload)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
