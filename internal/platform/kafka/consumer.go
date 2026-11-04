package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Consumer struct {
	client *kgo.Client
}

func NewConsumer(broker, group, topic string) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(broker),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka consumer: %w", err)
	}
	return &Consumer{client: client}, nil
}

func (c *Consumer) Run(ctx context.Context, handle func(ctx context.Context, value []byte) error, onFail func(ctx context.Context, value []byte, err error)) {
	for {
		fetches := c.client.PollFetches(ctx)
		if ctx.Err() != nil {
			return
		}
		for _, fe := range fetches.Errors() {
			log.Println("kafka fetch error:", fe.Err)
		}

		fetches.EachRecord(func(record *kgo.Record) {
			err := handleWithRetry(ctx, handle, record.Value)
			if ctx.Err() != nil {
				return
			}

			if err != nil {
				log.Println("giving up on message:", err)
				onFail(ctx, record.Value, err)
			}

			err = c.client.CommitRecords(ctx, record)
			if err != nil {
				log.Println("kafka commit error:", err)
			}
		})
	}
}

func (c *Consumer) Close() {
	c.client.Close()
}

func handleWithRetry(ctx context.Context, handle func(ctx context.Context, value []byte) error, value []byte) error {
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		err = handle(ctx, value)
		if err == nil {
			return nil
		}
		log.Printf("attempt %d failed: %v", attempt, err)
		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	return err
}
