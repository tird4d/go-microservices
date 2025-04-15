package config

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func ConnectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"), // z. B. "localhost:6379"
		Password: "",                      // kein Passwort standardmäßig
		DB:       0,                       // Default DB
	})
}
