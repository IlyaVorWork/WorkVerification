package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
	"workVerification/internal/adapters/adb"
	"workVerification/internal/adapters/kafka"
	"workVerification/internal/adapters/s3"
	"workVerification/internal/core/service"

	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
)

var responsesProducer *kafka.Producer
var verificationService *service.Service

func main() {
	minioEndpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "admin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "password123")
	minioBucket := getEnv("MINIO_BUCKET", "app-builds")
	kafkaBroker := getEnv("KAFKA_BROKER", "localhost:9092")
	emulatorHost := getEnv("EMULATOR_HOST", "localhost")
	emulatorPort := getEnv("EMULATOR_PORT", "5555")

	minioClient := s3.NewMinio(minioBucket, minioEndpoint, minioAccessKey, minioSecretKey)
	adbClient := adb.NewADB(emulatorHost, emulatorPort)

	verificationService = service.NewService(minioClient, adbClient)

	responsesProducer = kafka.NewProducer([]string{kafkaBroker})
	defer func() {
		if err := responsesProducer.Close(); err != nil {
			log.Printf("producer close error: %v", err)
		}
	}()

	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: []string{kafkaBroker},
		GroupID: "work.verify.request",
		Topic:   kafka.TopicWorkVerifyRequest,
	}, handleVerificationRequest)

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("consumer close error: %v", err)
		}
	}()

	if err := consumer.Run(ctx); err != nil && ctx.Err() == nil {
		log.Fatalf("consumer error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return fallback
}

func handleVerificationRequest(ctx context.Context, msg kafkago.Message) error {
	var req kafka.Event
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		return err
	}

	resp := kafka.Event{
		EventID:       uuid.NewString(),
		CorrelationID: req.CorrelationID,
		Type:          "work.verify.succeeded",
		Timestamp:     time.Now().UTC(),
	}
	log.Printf("verification request accepted: %v", msg)

	if err := verificationService.Verify(req.FileName); err != nil {
		resp.Type = "work.verify.failed"
		log.Printf("verification failed for %s: %v", req.CorrelationID, err)
	}

	log.Printf("verification finished: correlationID=%s type=%s", req.CorrelationID, resp.Type)
	return responsesProducer.Send(ctx, kafka.TopicVerificationResponses, req.CorrelationID, resp)
}
