package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tird4d/go-microservices/order_service/clients"
	"github.com/tird4d/go-microservices/order_service/consumer"
	"github.com/tird4d/go-microservices/order_service/events"
	"github.com/tird4d/go-microservices/order_service/handlers"
	"github.com/tird4d/go-microservices/order_service/metrics"
	"github.com/tird4d/go-microservices/order_service/storage"
	"github.com/tird4d/go-microservices/order_service/tracing"
)

func main() {
	_ = godotenv.Load()

	// Metrics
	metrics.InitMetrics()
	go func() {
		http.Handle("/metrics", promhttp.Handler())
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
	tp, err := tracing.InitTracer("order-service", jaegerEndpoint)
	if err != nil {
		log.Printf("⚠️ Failed to init tracer: %v", err)
	} else {
		defer tracing.Shutdown(tp)
		log.Printf("✅ Tracing initialized → %s", jaegerEndpoint)
	}

	// MongoDB
	storage.InitMongo()

	// RabbitMQ
	rabbitAddr := os.Getenv("RABBITMQ_URL")
	if rabbitAddr == "" {
		rabbitAddr = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitAddr)
	if err != nil {
		log.Fatalf("❌ Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Channel for consuming user events
	chConsume, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Failed to open consumer channel: %v", err)
	}
	defer chConsume.Close()

	// Channel for publishing order events
	chPublish, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Failed to open publisher channel: %v", err)
	}
	defer chPublish.Close()

	// Start user event consumer (saves users to memory)
	consumer.StartUserConsumer(chConsume)

	// Product gRPC client
	productAddr := os.Getenv("PRODUCT_SERVICE_ADDR")
	if productAddr == "" {
		productAddr = "localhost:50053"
	}
	productClient := clients.NewProductClient(productAddr)

	// Order event publisher
	publisher := events.NewOrderPublisher(chPublish)

	// HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	r := gin.Default()

	orderHandler := &handlers.OrderHandler{
		ProductClient: productClient,
		Publisher:     publisher,
	}

	v1 := r.Group("/api/v1")
	{
		v1.POST("/orders", orderHandler.CreateOrder)
		v1.GET("/orders", orderHandler.ListOrders)
	}

	log.Printf("🚀 order-service listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
