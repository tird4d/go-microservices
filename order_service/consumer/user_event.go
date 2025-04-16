package consumer

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tird4d/go-microservices/order_service/storage"
)

type UserRegisteredEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

func StartUserConsumer(ch *amqp.Channel) {
	err := ch.ExchangeDeclare("user_exchange", "fanout", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ Failed to declare exchange: %v", err)
	}

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		log.Fatalf("❌ Failed to declare queue: %v", err)
	}

	err = ch.QueueBind(q.Name, "", "user_exchange", false, nil)
	if err != nil {
		log.Fatalf("❌ Failed to bind queue: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, true, false, false, nil)
	if err != nil {
		log.Fatalf("❌ Failed to register consumer: %v", err)
	}

	log.Println("📡 Listening for user.created events...")

	go func() {
		for msg := range msgs {
			var event UserRegisteredEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Println("⚠️ Invalid message:", err)
				continue
			}

			storage.SaveUser(storage.User{
				ID:    event.UserID,
				Email: event.Email,
				Name:  event.Name,
			})

			log.Printf("✅ User %s saved (%s)", event.Name, event.UserID)
		}
	}()
}
