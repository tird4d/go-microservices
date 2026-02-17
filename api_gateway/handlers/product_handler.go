package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	productpb "github.com/tird4d/go-microservices/product_service/proto"
)

type ProductHandler struct {
	ProductClient productpb.ProductServiceClient
}

// CreateHandler handles HTTP POST /products - creates a new product
func (p *ProductHandler) CreateHandler(c *gin.Context) {
	var body struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Price       float64 `json:"price" binding:"required"`
		Category    string  `json:"category" binding:"required"`
		Stock       int32   `json:"stock"`
		ImageUrl    string  `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create gRPC context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call product-service via gRPC
	res, err := p.ProductClient.CreateProduct(ctx, &productpb.CreateProductRequest{
		Name:        body.Name,
		Description: body.Description,
		Price:       body.Price,
		Category:    body.Category,
		Stock:       body.Stock,
		ImageUrl:    body.ImageUrl,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return created product
	c.JSON(http.StatusCreated, gin.H{
		"id":          res.Id,
		"name":        res.Name,
		"description": res.Description,
		"price":       res.Price,
		"category":    res.Category,
		"stock":       res.Stock,
		"image_url":   res.ImageUrl,
		"created_at":  res.CreatedAt.AsTime(),
		"updated_at":  res.UpdatedAt.AsTime(),
	})
}

// GetProductHandler handles HTTP GET /products/:id - retrieves a product by ID
func (p *ProductHandler) GetProductHandler(c *gin.Context) {
	// Extract ID from URL parameter
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product ID is required"})
		return
	}

	// Create gRPC context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call product-service via gRPC
	res, err := p.ProductClient.GetProduct(ctx, &productpb.GetProductRequest{
		Id: productID,
	})

	if err != nil {
		// gRPC errors are translated here
		// NotFound -> 404, Internal -> 500, etc.
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	// Return product details
	c.JSON(http.StatusOK, gin.H{
		"id":          res.Id,
		"name":        res.Name,
		"description": res.Description,
		"price":       res.Price,
		"category":    res.Category,
		"stock":       res.Stock,
		"image_url":   res.ImageUrl,
		"created_at":  res.CreatedAt.AsTime(),
		"updated_at":  res.UpdatedAt.AsTime(),
	})
}

// UpdateProductHandler handles HTTP PUT /products/:id - updates an existing product
func (p *ProductHandler) UpdateProductHandler(c *gin.Context) {
	// Extract ID from URL parameter
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product ID is required"})
		return
	}

	// Bind request body
	var body struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
		Stock       int32   `json:"stock"`
		ImageUrl    string  `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create gRPC context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call product-service via gRPC
	// Note: Only fields that are provided will be updated
	res, err := p.ProductClient.UpdateProduct(ctx, &productpb.UpdateProductRequest{
		Id:          productID,
		Name:        body.Name,
		Description: body.Description,
		Price:       body.Price,
		Category:    body.Category,
		Stock:       body.Stock,
		ImageUrl:    body.ImageUrl,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated product
	c.JSON(http.StatusOK, gin.H{
		"id":          res.Id,
		"name":        res.Name,
		"description": res.Description,
		"price":       res.Price,
		"category":    res.Category,
		"stock":       res.Stock,
		"image_url":   res.ImageUrl,
		"created_at":  res.CreatedAt.AsTime(),
		"updated_at":  res.UpdatedAt.AsTime(),
	})
}

// DeleteProductHandler handles HTTP DELETE /products/:id - deletes a product
func (p *ProductHandler) DeleteProductHandler(c *gin.Context) {
	// Extract ID from URL parameter
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product ID is required"})
		return
	}

	// Create gRPC context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call product-service via gRPC
	_, err := p.ProductClient.DeleteProduct(ctx, &productpb.DeleteProductRequest{
		Id: productID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return success message
	c.JSON(http.StatusOK, gin.H{
		"message": "product deleted successfully",
	})
}

// ListProductsHandler handles HTTP GET /products - lists products with pagination
func (p *ProductHandler) ListProductsHandler(c *gin.Context) {
	// Get pagination parameters from query string
	// Example: /products?page=2&page_size=10
	var page int32 = 1
	var pageSize int32 = 10

	// Simple query parameter parsing with defaults
	if pageStr := c.DefaultQuery("page", "1"); pageStr != "" {
		if pageVal := parseIntDefault(pageStr, 1); pageVal > 0 {
			page = int32(pageVal)
		}
	}

	if sizeStr := c.DefaultQuery("page_size", "10"); sizeStr != "" {
		if sizeVal := parseIntDefault(sizeStr, 10); sizeVal > 0 {
			pageSize = int32(sizeVal)
		}
	}

	// Create gRPC context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call product-service via gRPC
	res, err := p.ProductClient.ListProducts(ctx, &productpb.ListProductsRequest{
		Page:     page,
		PageSize: pageSize,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert products to JSON-friendly format
	products := make([]gin.H, len(res.Products))
	for i, product := range res.Products {
		products[i] = gin.H{
			"id":          product.Id,
			"name":        product.Name,
			"description": product.Description,
			"price":       product.Price,
			"category":    product.Category,
			"stock":       product.Stock,
			"image_url":   product.ImageUrl,
			"created_at":  product.CreatedAt.AsTime(),
			"updated_at":  product.UpdatedAt.AsTime(),
		}
	}

	// Return paginated response
	c.JSON(http.StatusOK, gin.H{
		"products":    products,
		"total":       res.Total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (res.Total + pageSize - 1) / pageSize, // Ceiling division
	})
}

// GetProductsByCategoryHandler handles HTTP GET /products/category/:category - lists products by category
func (p *ProductHandler) GetProductsByCategoryHandler(c *gin.Context) {
	// Extract category from URL parameter
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category is required"})
		return
	}

	// Get pagination parameters (same as ListProducts)
	var page int32 = 1
	var pageSize int32 = 10

	if pageStr := c.DefaultQuery("page", "1"); pageStr != "" {
		if pageVal := parseIntDefault(pageStr, 1); pageVal > 0 {
			page = int32(pageVal)
		}
	}

	if sizeStr := c.DefaultQuery("page_size", "10"); sizeStr != "" {
		if sizeVal := parseIntDefault(sizeStr, 10); sizeVal > 0 {
			pageSize = int32(sizeVal)
		}
	}

	// Create gRPC context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call product-service via gRPC
	res, err := p.ProductClient.GetProductsByCategory(ctx, &productpb.GetProductsByCategoryRequest{
		Category: category,
		Page:     page,
		PageSize: pageSize,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert products to JSON-friendly format
	products := make([]gin.H, len(res.Products))
	for i, product := range res.Products {
		products[i] = gin.H{
			"id":          product.Id,
			"name":        product.Name,
			"description": product.Description,
			"price":       product.Price,
			"category":    product.Category,
			"stock":       product.Stock,
			"image_url":   product.ImageUrl,
			"created_at":  product.CreatedAt.AsTime(),
			"updated_at":  product.UpdatedAt.AsTime(),
		}
	}

	// Return filtered and paginated response
	c.JSON(http.StatusOK, gin.H{
		"products":    products,
		"category":    category,
		"total":       res.Total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (res.Total + pageSize - 1) / pageSize,
	})
}

// Helper function to parse int with default value
func parseIntDefault(s string, defaultVal int) int {
	var val int
	if _, err := fmt.Sscanf(s, "%d", &val); err == nil {
		return val
	}
	return defaultVal
}
