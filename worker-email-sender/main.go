package main

import (
	"context"
	"errors"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"log"
	"os"
	"time"
	"worker-email-sender/tasks"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if present
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing...")
	}

	// Get NATS URL from environment variables
	natsURL, found := os.LookupEnv("NATS_URL")
	if !found {
		panic("NATS_URL not found in environment variables")
	}

	consumerName := "email-consumer"

	// Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("could not connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Connected to NATS:", natsURL)

	jsm, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("could not connect to JetStream: %v", err)
	}

	streamConfig := jetstream.StreamConfig{
		Name:      "EMAIL_EVENTS",
		Subjects:  []string{tasks.TypeEmailDelivery, tasks.TypeEmailDeliveryAttachment},
		Retention: jetstream.WorkQueuePolicy,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := jsm.CreateOrUpdateStream(ctx, streamConfig)
	if err != nil && !errors.Is(err, jetstream.ErrStreamNameAlreadyInUse) {
		log.Fatalf("could not create stream: %v", err)
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:           consumerName,
		Durable:        consumerName,
		AckPolicy:      jetstream.AckExplicitPolicy, // Acknowledge each message explicitly
		DeliverPolicy:  jetstream.DeliverAllPolicy,  // Deliver all available messages
		FilterSubjects: streamConfig.Subjects,
		MaxDeliver:     5,                // Retry up to 5 times
		AckWait:        60 * time.Second, // Time to wait for acknowledgment
	})

	if err != nil && !errors.Is(err, nats.ErrConsumerNameAlreadyInUse) {
		log.Fatalf("could not create consumer: %v", err)
	}

	_, err = consumer.Consume(func(msg jetstream.Msg) {
		switch msg.Subject() {
		case tasks.TypeEmailDelivery:
			tasks.HandleEmailDeliveryMessage(msg)
		case tasks.TypeEmailDeliveryAttachment:
			tasks.HandleEmailAttachmentDeliveryMessage(msg)
		}
	})
	if err != nil {
		log.Fatalf("could not subscribe to email consumer: %v", err)
	}

	// Keep the connection alive
	select {}
}
