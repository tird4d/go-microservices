package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tird4d/go-microservices/order_service/consumer"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("❌ Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Failed to open channel: %v", err)
	}
	defer ch.Close()

	consumer.StartUserConsumer(ch)

	select {} // läuft endlos
}
