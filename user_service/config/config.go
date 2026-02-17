package config

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/tird4d/go-microservices/user_service/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB() (*mongo.Database, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal("MongoDB connection error: ", err)
		return nil, err
	}
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB ping error: ", err)
		return nil, err
	}

	DB = client.Database(os.Getenv("MONGO_DB"))

	logger.Log.Info("✅ MongoDB connected", "db", os.Getenv("MONGO_DB"))

	return DB, nil
}
