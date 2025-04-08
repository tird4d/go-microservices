package repositories

import (
	"github.com/tird4d/go-microservices/user_service/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	InsertNewUser(user *models.User) (*mongo.InsertOneResult, error)
}
