package services

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tird4d/go-microservices/product_service/logger"
	"github.com/tird4d/go-microservices/product_service/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")
	logger.InitLogger(true)

	if err != nil {
		logger.Log.Info("Error loading .env file")
	}

	os.Exit(m.Run())
}

// MockProductRepository is a mock implementation of MongoProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(product *models.Product) (*models.Product, error) {
	args := m.Called(product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, id primitive.ObjectID, updates map[string]any) (*models.Product, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) FindAll(ctx context.Context, skip, limit int64) ([]*models.Product, error) {
	args := m.Called(ctx, skip, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) FindByCategory(ctx context.Context, category string, skip, limit int64) ([]*models.Product, error) {
	args := m.Called(ctx, category, skip, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProductRepository) CountByCategory(ctx context.Context, category string) (int64, error) {
	args := m.Called(ctx, category)
	return args.Get(0).(int64), args.Error(1)
}

// --- CREATE PRODUCT TESTS ---

func TestCreateProduct_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	product := &models.Product{
		ID:          primitive.NewObjectID(),
		Name:        "Test Laptop",
		Description: "High-performance laptop",
		Price:       1299.99,
		Category:    "electronics",
		Stock:       10,
		ImageURL:    "https://example.com/laptop.jpg",
	}

	mockRepo.On("Create", mock.MatchedBy(func(p *models.Product) bool {
		return p.Name == "Test Laptop" && p.Price == 1299.99
	})).Return(product, nil)

	result, err := CreateProduct(ctx, mockRepo, "Test Laptop", "High-performance laptop", "electronics", "https://example.com/laptop.jpg", 1299.99, 10)

	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Laptop", result.Name)
	assert.Equal(t, int32(10), result.Stock)
}

func TestCreateProduct_Error(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	mockRepo.On("Create", mock.Anything).Return(nil, errors.New("database error"))

	result, err := CreateProduct(ctx, mockRepo, "Test Laptop", "High-performance laptop", "electronics", "https://example.com/laptop.jpg", 1299.99, 10)

	mockRepo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- GET PRODUCT BY ID TESTS ---

func TestGetProductByID_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	id := primitive.NewObjectID()
	product := &models.Product{
		ID:       id,
		Name:     "Test Laptop",
		Price:    1299.99,
		Stock:    10,
		Category: "electronics",
	}

	mockRepo.On("FindByID", ctx, id).Return(product, nil)

	result, err := GetProductByID(ctx, mockRepo, id.Hex())

	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Laptop", result.Name)
}

func TestGetProductByID_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	id := primitive.NewObjectID()
	mockRepo.On("FindByID", ctx, id).Return(nil, mongo.ErrNoDocuments)

	result, err := GetProductByID(ctx, mockRepo, id.Hex())

	mockRepo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetProductByID_InvalidIDFormat(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	result, err := GetProductByID(ctx, mockRepo, "invalid-id")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetProductByID_DatabaseError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	id := primitive.NewObjectID()
	mockRepo.On("FindByID", ctx, id).Return(nil, errors.New("connection error"))

	result, err := GetProductByID(ctx, mockRepo, id.Hex())

	mockRepo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- UPDATE PRODUCT TESTS ---

func TestUpdateProduct_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	id := primitive.NewObjectID()
	updatedProduct := &models.Product{
		ID:       id,
		Name:     "Updated Laptop",
		Price:    1399.99,
		Stock:    15,
		Category: "electronics",
	}

	mockRepo.On("Update", ctx, id, mock.MatchedBy(func(updates map[string]any) bool {
		return updates["name"] == "Updated Laptop" && updates["price"] == 1399.99
	})).Return(updatedProduct, nil)

	result, err := UpdateProduct(ctx, mockRepo, id.Hex(), "Updated Laptop", "High-performance laptop", "electronics", "https://example.com/laptop.jpg", 1399.99, 15)

	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Laptop", result.Name)
	assert.Equal(t, 1399.99, result.Price)
}

func TestUpdateProduct_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	id := primitive.NewObjectID()
	mockRepo.On("Update", ctx, id, mock.Anything).Return(nil, mongo.ErrNoDocuments)

	result, err := UpdateProduct(ctx, mockRepo, id.Hex(), "Updated Laptop", "Description", "electronics", "url", 1399.99, 15)

	mockRepo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUpdateProduct_InvalidIDFormat(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	result, err := UpdateProduct(ctx, mockRepo, "invalid-id", "Updated Laptop", "Description", "electronics", "url", 1399.99, 15)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- DELETE PRODUCT TESTS ---

func TestDeleteProduct_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	id := primitive.NewObjectID()
	mockRepo.On("Delete", ctx, id).Return(nil)

	err := DeleteProduct(ctx, mockRepo, id.Hex())

	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
}

func TestDeleteProduct_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	id := primitive.NewObjectID()
	mockRepo.On("Delete", ctx, id).Return(mongo.ErrNoDocuments)

	err := DeleteProduct(ctx, mockRepo, id.Hex())

	mockRepo.AssertExpectations(t)
	assert.Error(t, err)
}

func TestDeleteProduct_InvalidIDFormat(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	err := DeleteProduct(ctx, mockRepo, "invalid-id")

	assert.Error(t, err)
}

// --- LIST PRODUCTS TESTS (with pagination) ---

func TestListProducts_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	products := []*models.Product{
		{
			ID:    primitive.NewObjectID(),
			Name:  "Laptop",
			Price: 1299.99,
		},
		{
			ID:    primitive.NewObjectID(),
			Name:  "Mouse",
			Price: 29.99,
		},
	}

	mockRepo.On("FindAll", ctx, int64(0), int64(10)).Return(products, nil)
	mockRepo.On("Count", ctx).Return(int64(2), nil)

	result, total, err := ListProducts(ctx, mockRepo, 1, 10)

	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
}

func TestListProducts_Page2(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	products := []*models.Product{
		{
			ID:    primitive.NewObjectID(),
			Name:  "Item3",
			Price: 99.99,
		},
	}

	mockRepo.On("FindAll", ctx, int64(10), int64(10)).Return(products, nil)
	mockRepo.On("Count", ctx).Return(int64(11), nil)

	result, total, err := ListProducts(ctx, mockRepo, 2, 10)

	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, int64(11), total)
}

func TestListProducts_Empty(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	mockRepo.On("FindAll", ctx, int64(0), int64(10)).Return([]*models.Product{}, nil)
	mockRepo.On("Count", ctx).Return(int64(0), nil)

	result, total, err := ListProducts(ctx, mockRepo, 1, 10)

	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
}

func TestListProducts_DatabaseError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	mockRepo.On("FindAll", ctx, int64(0), int64(10)).Return(nil, errors.New("database error"))

	result, total, err := ListProducts(ctx, mockRepo, 1, 10)

	mockRepo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
}

// --- GET PRODUCTS BY CATEGORY TESTS (with pagination) ---

func TestGetProductsByCategory_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	category := "electronics"
	products := []*models.Product{
		{
			ID:       primitive.NewObjectID(),
			Name:     "Laptop",
			Category: "electronics",
			Price:    1299.99,
		},
		{
			ID:       primitive.NewObjectID(),
			Name:     "Mouse",
			Category: "electronics",
			Price:    29.99,
		},
	}

	mockRepo.On("FindByCategory", ctx, category, int64(0), int64(10)).Return(products, nil)
	mockRepo.On("CountByCategory", ctx, category).Return(int64(2), nil)

	result, total, err := GetProductsByCategory(ctx, mockRepo, category, 1, 10)

	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	assert.Equal(t, category, result[0].Category)
}

func TestGetProductsByCategory_Empty(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	category := "nonexistent"
	mockRepo.On("FindByCategory", ctx, category, int64(0), int64(10)).Return([]*models.Product{}, nil)
	mockRepo.On("CountByCategory", ctx, category).Return(int64(0), nil)

	result, total, err := GetProductsByCategory(ctx, mockRepo, category, 1, 10)

	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
}

func TestGetProductsByCategory_DatabaseError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductRepository)

	category := "electronics"
	mockRepo.On("FindByCategory", ctx, category, int64(0), int64(10)).Return(nil, errors.New("database error"))

	result, total, err := GetProductsByCategory(ctx, mockRepo, category, 1, 10)

	mockRepo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
}
