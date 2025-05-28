package repositories

import (
	"context"
	"time"

	"github.com/tird4d/go-microservices/user_service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	if err := models.UserCollection().FindOne(ctx, bson.M{"email": email}).Decode(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *MongoUserRepository) FindUserByID(ctx context.Context, oid primitive.ObjectID) (*models.User, error) {
	user := &models.User{}

	if err := models.UserCollection().FindOne(ctx, bson.M{"_id": oid}).Decode(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *MongoUserRepository) FindUsers(ctx context.Context, skip, pageSize int64) ([]*models.User, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize))

	cursor, err := models.UserCollection().Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err == nil {
			users = append(users, &user)
		}
	}

	return users, nil
}

func (r *MongoUserRepository) CountUsers(ctx context.Context) (int64, error) {

	return models.UserCollection().CountDocuments(ctx, bson.M{})
}

func (r *MongoUserRepository) UpdateUser(ctx context.Context, oid primitive.ObjectID, updates map[string]any) (*mongo.UpdateResult, error) {

	filter := bson.M{"_id": oid}
	updateFields := bson.M{"$set": updates}

	return models.UserCollection().UpdateOne(ctx, filter, updateFields)
}

func (r *MongoUserRepository) DeleteUser(ctx context.Context, oid primitive.ObjectID) (*mongo.DeleteResult, error) {
	filter := bson.M{"_id": oid}
	return models.UserCollection().DeleteOne(ctx, filter)
}
