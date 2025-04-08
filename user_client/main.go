package main

import (
	"context"
	"log"
	"time"

	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"google.golang.org/grpc"
)

func main() {
	// اتصال به gRPC سرور
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ could not connect: %v", err)
	}
	defer conn.Close()

	client := userpb.NewUserServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// ساخت request
	req := &userpb.RegisterRequest{
		Name:     "Ali",
		Email:    "ali@example.com",
		Password: "secret",
	}

	// ارسال درخواست
	res, err := client.Register(ctx, req)
	if err != nil {
		log.Fatalf("❌ Register failed: %v", err)
	}

	log.Printf("✅ Response: ID=%s, Message=%s", res.Id, res.Message)
}
