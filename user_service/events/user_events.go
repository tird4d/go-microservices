package events

import (
	"encoding/json"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type UserRegisteredEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

func PublishUserRegisteredEvent(event UserRegisteredEvent) error {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_CONNECTION_STRING"))
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"user_exchange", // Name
		"fanout",        // Typ
		true,            // Durable
		false,           // Auto-Delete
		false,           // Internal
		false,           // No-Wait
		nil,             // Args
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = ch.Publish(
		"user_exchange", // Exchange
		"",              // Routing Key
		false,           // Mandatory
		false,           // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return err
	}

	log.Println("📤 Event published:", string(body))
	return nil
}
