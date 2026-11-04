package kafka

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client *kgo.Client
}

func NewProducer(broker string) (*Producer, error) {
	client, err := kgo.NewClient(kgo.SeedBrokers(broker))
	if err != nil {
		return nil, fmt.Errorf("Create kafka client: %w", err)
	}
	return &Producer{client: client}, nil
}

func (p *Producer) Publish(ctx context.Context, topic, key string, value []byte) error {
	record := &kgo.Record{Topic: topic, Key: []byte(key), Value: value}
	return p.client.ProduceSync(ctx, record).FirstErr()
}

func (p *Producer) Close() {
	p.client.Close()
}
