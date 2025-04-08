package repositories

import (
	"context"
	"time"

	"github.com/tird4d/go-microservices/user_service/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoUserRepository struct{}

func (r *MongoUserRepository) InsertNewUser(user *models.User) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	result, err := models.UserCollection().InsertOne(ctx, user)

	defer cancel()

	return result, err
}
