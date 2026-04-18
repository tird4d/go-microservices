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

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int32   `json:"quantity"`
}

type OrderPlacedEvent struct {
	OrderID    string      `json:"order_id"`
	UserID     string      `json:"user_id"`
	UserEmail  string      `json:"user_email"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
}

func bindFanout(conn *amqp.Connection, exchange string) (<-chan amqp.Delivery, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err = ch.ExchangeDeclare(exchange, "fanout", true, false, false, false, nil); err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return nil, err
	}

	if err = ch.QueueBind(q.Name, "", exchange, false, nil); err != nil {
		return nil, err
	}

	return ch.Consume(q.Name, "", true, true, false, false, nil)
}

func main() {
	_ = godotenv.Load()

	rabbitMQAddr := os.Getenv("RABBITMQ_CONNECTION_STRING")
	if rabbitMQAddr == "" {
		rabbitMQAddr = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitMQAddr)
	if err != nil {
		log.Fatalf("❌ Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Consumer 1: user.registered → welcome email
	userMsgs, err := bindFanout(conn, "user_exchange")
	if err != nil {
		log.Fatalf("❌ Failed to bind user_exchange: %v", err)
	}

	// Consumer 2: order.placed → order confirmation email
	orderMsgs, err := bindFanout(conn, "order_exchange")
	if err != nil {
		log.Fatalf("❌ Failed to bind order_exchange: %v", err)
	}

	log.Println("📩 email-service waiting for events...")

	go func() {
		for msg := range userMsgs {
			var event UserRegisteredEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Println("⚠️ Failed to parse user event:", err)
				continue
			}
			log.Printf("📨 Welcome email → %s <%s>", event.Name, event.Email)
		}
	}()

	for msg := range orderMsgs {
		var event OrderPlacedEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Println("⚠️ Failed to parse order event:", err)
			continue
		}
		log.Printf("📦 Order confirmation → %s | order=%s | total=€%.2f (%d item(s))",
			event.UserEmail, event.OrderID, event.TotalPrice, len(event.Items))
	}
}
