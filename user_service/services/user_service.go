package services

import (
	"context"

	"github.com/tird4d/go-microservices/user_service/models"
	"github.com/tird4d/go-microservices/user_service/repositories"
	"github.com/tird4d/go-microservices/user_service/utils"
)

func RegisterUser(ctx context.Context, repo repositories.UserRepository, name, email, password string) error {

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	user := models.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Role:     "user",
	}

	_, err = repo.InsertNewUser(&user)

	return err
}

func GetUserCredential(ctx context.Context, repo repositories.UserRepository, email string) (*models.User, error) {
	user, err := repo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
