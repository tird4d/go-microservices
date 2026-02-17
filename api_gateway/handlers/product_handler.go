package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	productpb "github.com/tird4d/go-microservices/product_service/proto"
)

type ProductHandler struct {
	productpb.UnimplementedProductServiceServer
	ProductClient productpb.ProductServiceClient
}

func (p *ProductHandler) CreateHandler(c *gin.Context) {
	var body struct {
		Name        string    `bson:"name" json:"name"`
		Description string    `bson:"description" json:"description"`
		Price       float64   `bson:"price" json:"price"`
		Category    string    `bson:"category" json:"category"`
		Stock       int32     `bson:"stock" json:"stock"`
		ImageURL    string    `bson:"image_url" json:"image_url"`
		CreatedAt   time.Time `bson:"created_at" json:"created_at"`
		UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := p.ProductClient.Create(ctx, &productpb.RegisterRequest{
		Name:        body.Name,
		Description: body.Description,
		Price:       body.Price,
		Category:    body.Category,
		Stock:       body.Stock,
		ImageURL:    body.ImageURL,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": res.Id,
		"message": res.Message,
	})

}
