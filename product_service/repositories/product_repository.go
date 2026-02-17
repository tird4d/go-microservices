package repositories

import (
	"context"

	"github.com/tird4d/go-microservices/product_service/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	Create(product *models.Product) (*models.Product, error)
	FindByID(ctx context.Context, oid primitive.ObjectID) (*models.Product, error)
	FindAll(ctx context.Context, skip, pageSize int64) ([]*models.Product, error)
	FindByCategory(ctx context.Context, category string, skip, pageSize int64) ([]*models.Product, error)
	Count(ctx context.Context) (int64, error)
	CountByCategory(ctx context.Context, category string) (int64, error)
	Update(ctx context.Context, oid primitive.ObjectID, updates map[string]any) (*models.Product, error)
	Delete(ctx context.Context, oid primitive.ObjectID) error
}
