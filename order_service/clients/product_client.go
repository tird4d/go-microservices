package clients

import (
	"context"
	"log"
	"time"

	productpb "github.com/tird4d/go-microservices/product_service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProductClient struct {
	client productpb.ProductServiceClient
}

func NewProductClient(addr string) *ProductClient {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Failed to connect to product-service: %v", err)
	}
	return &ProductClient{client: productpb.NewProductServiceClient(conn)}
}

func (p *ProductClient) GetProduct(id string) (*productpb.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.client.GetProduct(ctx, &productpb.GetProductRequest{Id: id})
}
