package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tird4d/go-microservices/product_service/logger"
	"github.com/tird4d/go-microservices/product_service/models"
	"github.com/tird4d/go-microservices/product_service/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateProduct(ctx context.Context, repo *repositories.MongoProductRepository, name, description, category, imageUrl string, price float64, stock int32) (*models.Product, error) {

	product := models.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Category:    category,
		Stock:       stock,
		ImageURL:    imageUrl,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	createdProduct, err := repo.Create(&product)
	if err != nil {
		logger.Log.Errorw("Insert product failed", "error", err)
		return nil, fmt.Errorf("product insert failed: %w", err)
	}

	logger.Log.Infow("Product created successfully",
		"id", createdProduct.ID.Hex(),
		"name", createdProduct.Name,
	)

	return createdProduct, nil
}

func GetProductByID(ctx context.Context, repo *repositories.MongoProductRepository, id string) (*models.Product, error) {
	// Convert string ID to MongoDB ObjectID type
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Log.Errorw("Failed to convert string to ObjectID", "error", err)
		return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
	}

	product, err := repo.FindByID(ctx, oid)

	if err != nil {
		// Special handling: NotFound error should return 404
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		// Other errors are internal server errors
		logger.Log.Errorw("Failed to find product by ID", "error", err)
		return nil, status.Error(codes.Internal, "failed to retrieve product")
	}

	return product, nil
}

// UpdateProduct updates an existing product
// NOTE: UpdatedAt timestamp is set automatically in repository layer
func UpdateProduct(ctx context.Context, repo *repositories.MongoProductRepository, id, name, description, category, imageUrl string, price float64, stock int32) (*models.Product, error) {
	// Convert string ID to ObjectID
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Log.Errorw("Failed to convert string to ObjectID", "error", err)
		return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
	}

	// Build update map with only the fields to update
	// Repository pattern: updates map is database-agnostic
	updates := map[string]any{
		"name":        name,
		"description": description,
		"price":       price,
		"category":    category,
		"stock":       stock,
		"image_url":   imageUrl,
		// updated_at is set automatically by repository
	}

	// Repository handles the database update
	updatedProduct, err := repo.Update(ctx, oid, updates)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		logger.Log.Errorw("Failed to update product", "error", err)
		return nil, status.Error(codes.Internal, "failed to update product")
	}

	logger.Log.Infow("Product updated successfully",
		"id", updatedProduct.ID.Hex(),
		"name", updatedProduct.Name,
	)

	return updatedProduct, nil
}

// DeleteProduct removes a product from the database
func DeleteProduct(ctx context.Context, repo *repositories.MongoProductRepository, id string) error {
	// Convert string ID to ObjectID
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Log.Errorw("Failed to convert string to ObjectID", "error", err)
		return status.Error(codes.InvalidArgument, "invalid product ID format")
	}

	// Delete from database
	err = repo.Delete(ctx, oid)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return status.Error(codes.NotFound, "product not found")
		}
		logger.Log.Errorw("Failed to delete product", "error", err)
		return status.Error(codes.Internal, "failed to delete product")
	}

	logger.Log.Infow("Product deleted successfully", "id", id)
	return nil
}

// ListProducts retrieves all products with pagination
// PAGINATION PATTERN: skip = (page - 1) × pageSize
// Example: page=2, pageSize=10 → skip first 10, return next 10
func ListProducts(ctx context.Context, repo *repositories.MongoProductRepository, page, pageSize int32) ([]*models.Product, int64, error) {
	// Set defaults for pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10 // Default page size
	}

	// Calculate skip value: how many documents to skip
	// Page 1: skip 0, Page 2: skip pageSize, Page 3: skip pageSize*2
	skip := int64((page - 1) * pageSize)
	pageSizeInt64 := int64(pageSize)

	// Get products from database
	products, err := repo.FindAll(ctx, skip, pageSizeInt64)
	if err != nil {
		logger.Log.Errorw("Failed to list products", "error", err)
		return nil, 0, status.Error(codes.Internal, "failed to list products")
	}

	// Get total count for pagination metadata
	total, err := repo.Count(ctx)
	if err != nil {
		logger.Log.Errorw("Failed to count products", "error", err)
		return nil, 0, status.Error(codes.Internal, "failed to count products")
	}

	logger.Log.Infow("Products listed successfully",
		"total", total,
		"returned", len(products),
		"page", page,
	)

	return products, total, nil
}

// GetProductsByCategory retrieves products filtered by category with pagination
// SAME PAGINATION PATTERN as ListProducts, but with category filter
func GetProductsByCategory(ctx context.Context, repo *repositories.MongoProductRepository, category string, page, pageSize int32) ([]*models.Product, int64, error) {
	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Calculate skip for pagination - convert int32 to int64 for MongoDB driver
	skip := int64((page - 1) * pageSize)
	pageSizeInt64 := int64(pageSize)

	// Get filtered products
	products, err := repo.FindByCategory(ctx, category, skip, pageSizeInt64)
	if err != nil {
		logger.Log.Errorw("Failed to get products by category", "error", err, "category", category)
		return nil, 0, status.Error(codes.Internal, "failed to get products by category")
	}

	// Get total count for this category
	total, err := repo.CountByCategory(ctx, category)
	if err != nil {
		logger.Log.Errorw("Failed to count products by category", "error", err, "category", category)
		return nil, 0, status.Error(codes.Internal, "failed to count products")
	}

	logger.Log.Infow("Products by category retrieved successfully",
		"category", category,
		"total", total,
		"returned", len(products),
		"page", page,
	)

	return products, total, nil
}
