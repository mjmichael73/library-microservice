package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
)

func main() {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "broker:29092"
	}

	topics := []string{"user-registered-topic", "borrow-created-topic", "book-returned-topic"}
	for _, topic := range topics {
		go consume(context.Background(), broker, topic)
	}

	log.Printf("notificationservice started, consuming %v from %s", topics, broker)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("notificationservice shutting down")
}

func consume(ctx context.Context, broker, topic string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: "notificationservice",
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("notificationservice: read error on %s: %v", topic, err)
			return
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			payload = map[string]interface{}{"raw": string(msg.Value)}
		}
		log.Printf("notificationservice: received %s event: %+v", topic, payload)
		// Email delivery is intentionally a no-op placeholder until an SMTP
		// provider is configured; the event is logged for observability.
	}
}