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
