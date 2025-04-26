package server

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	brokerAddress string
}

func NewKafkaProducer(brokerAddress string) *KafkaProducer {
	return &KafkaProducer{brokerAddress: brokerAddress}
}

func (kp *KafkaProducer) SendMessage(ctx context.Context, topic string, key, value []byte) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kp.brokerAddress},
		Topic:   topic,
	})
	defer writer.Close()

	err := writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})

	if err != nil {
		log.Printf("failed to write message to Kafka: %v", err)
	}
	return err
}
