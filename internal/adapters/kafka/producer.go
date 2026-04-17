package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type Topic string

const (
	TopicWorkVerifyRequest Topic = "work.verify.request"

	TopicVerificationResponses Topic = "verification.responses"
)

type Event struct {
	EventID       string    `json:"EventID"`
	CorrelationID string    `json:"CorrelationID"`
	Type          string    `json:"Type"`
	FileName      string    `json:"FileName"`
	Timestamp     time.Time `json:"Timestamp"`
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			RequiredAcks: kafka.RequireAll,
		},
	}
}

func (p *Producer) Send(ctx context.Context, topic Topic, key string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: string(topic),
		Key:   []byte(key),
		Value: data,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
