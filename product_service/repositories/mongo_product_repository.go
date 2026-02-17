package repositories

import (
	"context"
	"time"

	"github.com/tird4d/go-microservices/product_service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoProductRepository struct{}

// Create inserts a new product and returns the created product with ID
func (r *MongoProductRepository) Create(product *models.Product) (*models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set timestamps
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	result, err := models.ProductCollection().InsertOne(ctx, product)
	if err != nil {
		return nil, err
	}

	// Set the generated ID
	product.ID = result.InsertedID.(primitive.ObjectID)

	return product, nil
}

// FindByID retrieves a product by ID
func (r *MongoProductRepository) FindByID(ctx context.Context, oid primitive.ObjectID) (*models.Product, error) {
	product := &models.Product{}

	if err := models.ProductCollection().FindOne(ctx, bson.M{"_id": oid}).Decode(product); err != nil {
		return nil, err
	}

	return product, nil
}

// FindAll retrieves products with pagination
func (r *MongoProductRepository) FindAll(ctx context.Context, skip, pageSize int64) ([]*models.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(pageSize)

	cursor, err := models.ProductCollection().Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*models.Product
	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err == nil {
			products = append(products, &product)
		}
	}

	return products, nil
}

// FindByCategory retrieves products by category with pagination
func (r *MongoProductRepository) FindByCategory(ctx context.Context, category string, skip, pageSize int64) ([]*models.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"category": category}
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(pageSize)

	cursor, err := models.ProductCollection().Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*models.Product
	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err == nil {
			products = append(products, &product)
		}
	}

	return products, nil
}

// Count returns total number of products
func (r *MongoProductRepository) Count(ctx context.Context) (int64, error) {
	return models.ProductCollection().CountDocuments(ctx, bson.M{})
}

// CountByCategory returns total number of products in a category
func (r *MongoProductRepository) CountByCategory(ctx context.Context, category string) (int64, error) {
	filter := bson.M{"category": category}
	return models.ProductCollection().CountDocuments(ctx, filter)
}

// Update updates a product and returns the updated product
func (r *MongoProductRepository) Update(ctx context.Context, oid primitive.ObjectID, updates map[string]any) (*models.Product, error) {
	// Add updated_at timestamp
	updates["updated_at"] = time.Now()

	filter := bson.M{"_id": oid}
	updateFields := bson.M{"$set": updates}

	// Use FindOneAndUpdate to return the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	product := &models.Product{}
	err := models.ProductCollection().FindOneAndUpdate(ctx, filter, updateFields, opts).Decode(product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

// Delete removes a product by ID
func (r *MongoProductRepository) Delete(ctx context.Context, oid primitive.ObjectID) error {
	filter := bson.M{"_id": oid}
	result, err := models.ProductCollection().DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// 	if err := models.UserCollection().FindOne(ctx, bson.M{"email": email}).Decode(user); err != nil {
// 		return nil, err
// 	}

// 	return user, nil
// }

// func (r *MongoUserRepository) FindUserByID(ctx context.Context, oid primitive.ObjectID) (*models.User, error) {
// 	user := &models.User{}

// 	if err := models.UserCollection().FindOne(ctx, bson.M{"_id": oid}).Decode(user); err != nil {
// 		return nil, err
// 	}

// 	return user, nil
// }

// func (r *MongoUserRepository) FindUsers(ctx context.Context, skip, pageSize int64) ([]*models.User, error) {

// 	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 	defer cancel()

// 	findOptions := options.Find().
// 		SetSkip(int64(skip)).
// 		SetLimit(int64(pageSize))

// 	cursor, err := models.UserCollection().Find(ctx, bson.M{}, findOptions)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cursor.Close(ctx)

// 	var users []*models.User
// 	for cursor.Next(ctx) {
// 		var user models.User
// 		if err := cursor.Decode(&user); err == nil {
// 			users = append(users, &user)
// 		}
// 	}

// 	return users, nil
// }

// func (r *MongoUserRepository) CountUsers(ctx context.Context) (int64, error) {

// 	return models.UserCollection().CountDocuments(ctx, bson.M{})
// }

// func (r *MongoUserRepository) UpdateUser(ctx context.Context, oid primitive.ObjectID, updates map[string]any) (*mongo.UpdateResult, error) {

// 	filter := bson.M{"_id": oid}
// 	updateFields := bson.M{"$set": updates}

// 	return models.UserCollection().UpdateOne(ctx, filter, updateFields)
// }

// func (r *MongoUserRepository) DeleteUser(ctx context.Context, oid primitive.ObjectID) (*mongo.DeleteResult, error) {
// 	filter := bson.M{"_id": oid}
// 	return models.UserCollection().DeleteOne(ctx, filter)
// }
