package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/tird4d/go-microservices/user_service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MongoUserRepository struct{}

func (r *MongoUserRepository) InsertNewUser(user *models.User) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	result, err := models.UserCollection().InsertOne(ctx, user)

	defer cancel()

	return result, err
}

func (r *MongoUserRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}

	err := models.UserCollection().FindOne(ctx, bson.M{"email": email}).Decode(user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, "user with this email not found")
		}
		return nil, err
	}

	return user, nil
}
