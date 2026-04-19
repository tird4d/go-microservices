package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tird4d/go-microservices/email_service/metrics"
	"github.com/tird4d/go-microservices/email_service/tracing"
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

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func processUserEvent(body []byte) error {
	var event UserRegisteredEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}
	log.Printf("📨 Welcome email → %s <%s>", event.Name, event.Email)
	metrics.EmailsSentCounter.WithLabelValues("welcome").Inc()
	return nil
}

func processOrderEvent(body []byte) error {
	var event OrderPlacedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}
	log.Printf("📦 Order confirmation → %s | order=%s | total=€%.2f (%d item(s))",
		event.UserEmail, event.OrderID, event.TotalPrice, len(event.Items))
	metrics.EmailsSentCounter.WithLabelValues("order_confirmation").Inc()
	return nil
}

func main() {
	_ = godotenv.Load()

	// Metrics + health
	metrics.InitMetrics()
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/health", healthHandler)
		log.Println("✅ Metrics server running on :2112")
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Fatalf("❌ Metrics server error: %v", err)
		}
	}()

	// Tracing
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "jaeger:4317"
	}
	tp, err := tracing.InitTracer("email-service", jaegerEndpoint)
	if err != nil {
		log.Printf("⚠️ Failed to init tracer: %v", err)
	} else {
		defer tracing.Shutdown(tp)
		log.Printf("✅ Tracing initialized → %s", jaegerEndpoint)
	}

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
			if err := processUserEvent(msg.Body); err != nil {
				log.Println("⚠️ Failed to parse user event:", err)
			}
		}
	}()

	for msg := range orderMsgs {
		if err := processOrderEvent(msg.Body); err != nil {
			log.Println("⚠️ Failed to parse order event:", err)
		}
	}
}
