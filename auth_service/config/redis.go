package config

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func ConnectRedis() {
	addr := os.Getenv("REDIS_ADDR")

	// Optional: enable TLS via env (recommended for ElastiCache Serverless)
	useTLS, _ := strconv.ParseBool(os.Getenv("REDIS_TLS")) // "true"/"false"

	opts := &redis.Options{
		Addr: addr,
		DB:   0,

		// Optional AUTH (if you enable access control later)
		Username: os.Getenv("REDIS_USERNAME"), // can be empty
		Password: os.Getenv("REDIS_PASSWORD"), // can be empty
	}

	if useTLS {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			// ServerName: "" // usually not needed; Go uses hostname from Addr for SNI
		}
	}

	RedisClient = redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis (addr=%s tls=%v): %v", addr, useTLS, err)
	}

	log.Println("✅ Redis client connected")
}
