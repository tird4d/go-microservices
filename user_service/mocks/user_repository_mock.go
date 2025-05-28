package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/tird4d/go-microservices/user_service/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) InsertNewUser(user *models.User) (*mongo.InsertOneResult, error) {
	args := m.Called(user)

	if result, ok := args.Get(0).(*mongo.InsertOneResult); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *UserRepositoryMock) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(email)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *UserRepositoryMock) FindUserByID(ctx context.Context, oid primitive.ObjectID) (*models.User, error) {
	args := m.Called(ctx, oid)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *UserRepositoryMock) FindUsers(ctx context.Context, skip, pageSize int64) ([]*models.User, error) {

	args := m.Called(ctx, skip, pageSize)
	if users, ok := args.Get(0).([]*models.User); ok {
		return users, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *UserRepositoryMock) CountUsers(ctx context.Context) (int64, error) {

	args := m.Called(ctx)
	if count, ok := args.Get(0).(int64); ok {
		return count, args.Error(1)
	}
	return 0, args.Error(1)
}

func (m *UserRepositoryMock) UpdateUser(ctx context.Context, oid primitive.ObjectID, updates map[string]any) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, oid, updates)
	if result, ok := args.Get(0).(*mongo.UpdateResult); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *UserRepositoryMock) DeleteUser(ctx context.Context, oid primitive.ObjectID) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, oid)
	if result, ok := args.Get(0).(*mongo.DeleteResult); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}
