package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/sashank33v/recallengine/internal/platform/kafka"
)

type DeadLetterPublisher struct {
	producer *kafka.Producer
	topic    string
}

func NewDeadLetterPublisher(producer *kafka.Producer, topic string) *DeadLetterPublisher {
	return &DeadLetterPublisher{producer: producer, topic: topic}
}

func (d *DeadLetterPublisher) Send(ctx context.Context, value []byte, cause error) {
	message, err := json.Marshal(map[string]string{
		"error":     cause.Error(),
		"message":   string(value),
		"failed_at": time.Now().UTC().Format(time.RFC3339),
	})

	if err != nil {
		log.Println("could not build dead-letter message:", err)
		return
	}

	err = d.producer.Publish(ctx, d.topic, "dlq", message)
	if err != nil {
		log.Println("could not write to dead-letter topic:", err)
		return
	}
	log.Println("message moved to", d.topic)
}
