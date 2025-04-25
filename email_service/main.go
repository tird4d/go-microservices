package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

type UserRegisteredEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("⚠️ Error loading .env file")
	}

	rabbitMQAddr := os.Getenv("RABBITMQ_CONNECTION_STRING")

	conn, err := amqp.Dial(rabbitMQAddr)
	if err != nil {
		log.Fatalf("❌ Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Failed to open channel: %v", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"user_exchange",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("❌ Failed to declare exchange: %v", err)
	}

	q, err := ch.QueueDeclare(
		"",    // zufälliger Queue-Name
		false, // nicht dauerhaft
		true,  // automatisch löschen
		true,  // exklusiv
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("❌ Failed to declare queue: %v", err)
	}

	err = ch.QueueBind(
		q.Name,
		"",
		"user_exchange",
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("❌ Failed to bind queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,  // auto-ack
		true,  // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("❌ Failed to register consumer: %v", err)
	}

	log.Println("📩 Waiting for events...")
	for msg := range msgs {
		var event UserRegisteredEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Println("⚠️ Failed to parse message:", err)
			continue
		}
		log.Printf("📨 New user registered! Sending welcome email to %s <%s>\n", event.Name, event.Email)
	}
}
