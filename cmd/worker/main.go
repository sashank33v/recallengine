package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sashank33v/recallengine/internal/config"
	"github.com/sashank33v/recallengine/internal/platform/kafka"
	"github.com/sashank33v/recallengine/internal/platform/postgres"
	"github.com/sashank33v/recallengine/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config error:", err)
	}
	db, err := postgres.ConnectDB(cfg)
	if err != nil {
		log.Fatal("database error:", err)
	}
	defer db.Close()

	producer, err := kafka.NewProducer(cfg.KafkaBroker)
	if err != nil {
		log.Fatal("Kafka error:", err)
	}
	defer producer.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	publisher := worker.NewOutboxPublisher(db, producer)
	consumer, err := kafka.NewConsumer(cfg.KafkaBroker, "mastery-service", "answer.events")
	if err != nil {
		log.Fatal("kafka consumer error:", err)
	}
	defer consumer.Close()

	mastery := worker.NewMasteryUpdater(db)
	dlq := worker.NewDeadLetterPublisher(producer, "answer.events.dlq")
	go consumer.Run(ctx, mastery.Handle, dlq.Send)

	log.Println("worker started...")
	publisher.Run(ctx)
	log.Println("worker stopped...")
}
