package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/tird4d/go-microservices/user_service/logger"
	"github.com/tird4d/go-microservices/user_service/models"
	"github.com/tird4d/go-microservices/user_service/repositories"
	"github.com/tird4d/go-microservices/user_service/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterUser(ctx context.Context, repo repositories.UserRepository, name, email, password string) (*mongo.InsertOneResult, error) {

	// Check if the email is already registered
	existingUser, err := repo.FindUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logger.Log.Errorw("Failed to check existing email", "error", err)
		return nil, fmt.Errorf("check existing user failed: %w", err)
	}

	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		logger.Log.Errorw("Failed to hash password", "error", err)
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	user := models.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Role:     "user",
	}

	result, err := repo.InsertNewUser(&user)
	if err != nil {
		logger.Log.Errorw("Insert user failed", "error", err)
		return nil, fmt.Errorf("user insert failed: %w", err)
	}

	return result, nil

}

func GetUserCredential(ctx context.Context, repo repositories.UserRepository, email string) (*models.User, error) {
	user, err := repo.FindUserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			logger.Log.Errorw("Failed to find user by email", "error", err)
		}
		return nil, err
	}
	return user, nil
}

func GetUserByID(ctx context.Context, repo repositories.UserRepository, id string) (*models.User, error) {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Log.Errorw("Failed to convert string to ObjectID", "error", err)
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	user, err := repo.FindUserByID(ctx, oid)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			logger.Log.Errorw("Failed to find user by ID", "error", err)
		}
		return nil, err
	}
	return user, nil
}
