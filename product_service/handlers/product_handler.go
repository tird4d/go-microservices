// product_handler.go
package handlers

import (
	"context"
	"fmt"

	"github.com/tird4d/go-microservices/product_service/logger"
	productpb "github.com/tird4d/go-microservices/product_service/proto"
	"github.com/tird4d/go-microservices/product_service/repositories"
	"github.com/tird4d/go-microservices/product_service/services"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	productpb.UnimplementedProductServiceServer
}

// Mock data storage (in-memory for now)
var mockProducts = make(map[string]*productpb.Product)
var mockProductList = []*productpb.Product{
	{
		Id:          "1",
		Name:        "Laptop Pro 15",
		Description: "High-performance laptop with 16GB RAM and 512GB SSD",
		Price:       1299.99,
		Category:    "electronics",
		Stock:       25,
		ImageUrl:    "https://example.com/laptop.jpg",
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
	},
	{
		Id:          "2",
		Name:        "Wireless Mouse",
		Description: "Ergonomic wireless mouse with 6 buttons",
		Price:       29.99,
		Category:    "electronics",
		Stock:       150,
		ImageUrl:    "https://example.com/mouse.jpg",
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
	},
	{
		Id:          "3",
		Name:        "Programming Book",
		Description: "Learn Go programming from scratch",
		Price:       45.00,
		Category:    "books",
		Stock:       80,
		ImageUrl:    "https://example.com/book.jpg",
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
	},
}

func init() {
	// Initialize mock data
	for _, product := range mockProductList {
		mockProducts[product.Id] = product
	}
}

// CreateProduct creates a new product
func (s *Server) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.Product, error) {
	logger.Log.Infow("Creating new product",
		"name", req.Name,
		"price", req.Price,
		"category", req.Category,
	)

	repo := &repositories.MongoProductRepository{}

	// Create product via service layer
	product, err := services.CreateProduct(
		ctx,
		repo,
		req.GetName(),
		req.GetDescription(),
		req.GetCategory(),
		req.GetImageUrl(),
		req.GetPrice(),
		req.GetStock(),
	)
	if err != nil {
		logger.Log.Errorw("Failed to create product", "error", err)
		return nil, err
	}

	// Convert model to protobuf response
	productResponse := &productpb.Product{
		Id:          product.ID.Hex(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Category:    product.Category,
		Stock:       product.Stock,
		ImageUrl:    product.ImageURL,
		CreatedAt:   timestamppb.New(product.CreatedAt),
		UpdatedAt:   timestamppb.New(product.UpdatedAt),
	}

	logger.Log.Infow("Product created successfully",
		"id", productResponse.Id,
		"name", productResponse.Name,
	)

	return productResponse, nil
}

// GetProduct retrieves a product by ID
func (s *Server) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.Product, error) {
	logger.Log.Infow("Getting product", "id", req.Id)

	repo := &repositories.MongoProductRepository{}

	// Convert string ID to ObjectID
	product, err := services.GetProductByID(ctx, repo, req.GetId())

	if err != nil {
		return nil, err
	}

	productResponse := &productpb.Product{
		Id:          product.ID.Hex(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Category:    product.Category,
		Stock:       product.Stock,
		ImageUrl:    product.ImageURL,
		CreatedAt:   timestamppb.New(product.CreatedAt),
		UpdatedAt:   timestamppb.New(product.UpdatedAt),
	}

	logger.Log.Infow("Product retrieved successfully", "id", product.ID)
	return productResponse, nil
}

// UpdateProduct updates an existing product
func (s *Server) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.Product, error) {
	logger.Log.Infow("Updating product", "id", req.Id)

	repo := &repositories.MongoProductRepository{}

	// Call service layer to update product
	product, err := services.UpdateProduct(
		ctx,
		repo,
		req.GetId(),
		req.GetName(),
		req.GetDescription(),
		req.GetCategory(),
		req.GetImageUrl(),
		req.GetPrice(),
		req.GetStock(),
	)
	if err != nil {
		logger.Log.Errorw("Failed to update product", "error", err)
		return nil, err
	}

	// Convert model to protobuf response
	productResponse := &productpb.Product{
		Id:          product.ID.Hex(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Category:    product.Category,
		Stock:       product.Stock,
		ImageUrl:    product.ImageURL,
		CreatedAt:   timestamppb.New(product.CreatedAt),
		UpdatedAt:   timestamppb.New(product.UpdatedAt),
	}

	logger.Log.Infow("Product updated successfully",
		"id", productResponse.Id,
		"name", productResponse.Name,
	)

	return productResponse, nil
}

// DeleteProduct deletes a product
func (s *Server) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*productpb.DeleteProductResponse, error) {
	logger.Log.Infow("Deleting product", "id", req.Id)

	repo := &repositories.MongoProductRepository{}

	// Call service layer to delete product
	err := services.DeleteProduct(ctx, repo, req.GetId())
	if err != nil {
		logger.Log.Errorw("Failed to delete product", "error", err)
		// Return structured error response
		return &productpb.DeleteProductResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete product: %v", err),
		}, err
	}

	logger.Log.Infow("Product deleted successfully", "id", req.Id)

	return &productpb.DeleteProductResponse{
		Success: true,
		Message: "Product deleted successfully",
	}, nil
}

// ListProducts lists all products with pagination
// PAGINATION: skip = (page-1) × pageSize is calculated in service layer
func (s *Server) ListProducts(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error) {
	logger.Log.Infow("Listing products",
		"page", req.Page,
		"pageSize", req.PageSize,
	)

	repo := &repositories.MongoProductRepository{}

	// Service layer handles pagination logic and database calls
	products, total, err := services.ListProducts(ctx, repo, req.GetPage(), req.GetPageSize())
	if err != nil {
		logger.Log.Errorw("Failed to list products", "error", err)
		return nil, err
	}

	// Convert domain models to protobuf responses
	productResponses := make([]*productpb.Product, len(products))
	for i, product := range products {
		productResponses[i] = &productpb.Product{
			Id:          product.ID.Hex(),
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Category:    product.Category,
			Stock:       product.Stock,
			ImageUrl:    product.ImageURL,
			CreatedAt:   timestamppb.New(product.CreatedAt),
			UpdatedAt:   timestamppb.New(product.UpdatedAt),
		}
	}

	// Use actual page values from request (service layer applies defaults)
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	logger.Log.Infow("Products listed successfully",
		"total", total,
		"returned", len(productResponses),
	)

	return &productpb.ListProductsResponse{
		Products: productResponses,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetProductsByCategory filters products by category with pagination
// SAME PATTERN as ListProducts, but with category filter in repository
func (s *Server) GetProductsByCategory(ctx context.Context, req *productpb.GetProductsByCategoryRequest) (*productpb.ListProductsResponse, error) {
	logger.Log.Infow("Getting products by category",
		"category", req.Category,
		"page", req.Page,
		"pageSize", req.PageSize,
	)

	repo := &repositories.MongoProductRepository{}

	// Service layer handles category filter + pagination
	products, total, err := services.GetProductsByCategory(
		ctx,
		repo,
		req.GetCategory(),
		req.GetPage(),
		req.GetPageSize(),
	)
	if err != nil {
		logger.Log.Errorw("Failed to get products by category", "error", err)
		return nil, err
	}

	// Convert domain models to protobuf responses
	productResponses := make([]*productpb.Product, len(products))
	for i, product := range products {
		productResponses[i] = &productpb.Product{
			Id:          product.ID.Hex(),
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Category:    product.Category,
			Stock:       product.Stock,
			ImageUrl:    product.ImageURL,
			CreatedAt:   timestamppb.New(product.CreatedAt),
			UpdatedAt:   timestamppb.New(product.UpdatedAt),
		}
	}

	// Apply default values for response metadata
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	logger.Log.Infow("Products by category retrieved successfully",
		"category", req.Category,
		"total", total,
		"returned", len(productResponses),
	)

	return &productpb.ListProductsResponse{
		Products: productResponses,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
