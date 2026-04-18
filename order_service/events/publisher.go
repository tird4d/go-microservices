package events

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tird4d/go-microservices/order_service/models"
)

type OrderPlacedEvent struct {
	OrderID    string            `json:"order_id"`
	UserID     string            `json:"user_id"`
	UserEmail  string            `json:"user_email"`
	Items      []models.OrderItem `json:"items"`
	TotalPrice float64           `json:"total_price"`
}

type OrderPublisher struct {
	ch *amqp.Channel
}

func NewOrderPublisher(ch *amqp.Channel) *OrderPublisher {
	err := ch.ExchangeDeclare(
		"order_exchange",
		"fanout",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("❌ Failed to declare order_exchange: %v", err)
	}
	return &OrderPublisher{ch: ch}
}

func (p *OrderPublisher) PublishOrderPlaced(order models.Order) {
	event := OrderPlacedEvent{
		OrderID:    order.ID,
		UserID:     order.UserID,
		UserEmail:  order.UserEmail,
		Items:      order.Items,
		TotalPrice: order.TotalPrice,
	}

	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("⚠️ Failed to marshal order event: %v", err)
		return
	}

	err = p.ch.Publish(
		"order_exchange", // exchange
		"",               // routing key (fanout ignores it)
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("⚠️ Failed to publish order.placed: %v", err)
		return
	}

	log.Printf("📤 Published order.placed for order %s (user: %s)", order.ID, order.UserEmail)
}
