package storage

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/tird4d/go-microservices/order_service/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var collection *mongo.Collection

func InitMongo() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "orders_db"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("❌ MongoDB connect error: %v", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("❌ MongoDB ping error: %v", err)
	}

	collection = client.Database(dbName).Collection("orders")
	log.Printf("✅ Connected to MongoDB Atlas (%s.orders)", dbName)
}

func SaveOrder(order models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, order)
	return err
}

func GetOrdersByUser(userID string) ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}
