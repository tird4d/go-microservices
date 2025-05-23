package services

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tird4d/go-microservices/user_service/logger"
	"github.com/tird4d/go-microservices/user_service/mocks"
	"github.com/tird4d/go-microservices/user_service/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")

	if err != nil {
		logger.Log.Error("Error loading .env file")
	}

	os.Exit(m.Run())

}

func TestRegisterUser_Success(t *testing.T) {

	//Context
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	id := primitive.NewObjectID()
	user := models.User{
		Name:     "test",
		Email:    "test@test.com",
		Password: "123456",
		Role:     "user",
	}

	mockRepo.On("FindUserByEmail", mock.MatchedBy(func(email string) bool {
		return true
	})).Return(nil, nil)

	mockRepo.On("InsertNewUser", mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "test@test.com" && u.Name == "test"
	})).Return(&mongo.InsertOneResult{InsertedID: id}, nil)

	result, err := RegisterUser(ctx, mockRepo, user.Name, user.Email, user.Password, user.Role)

	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, id, result.InsertedID)

}

func TestRegisterUser_UserAlreadyExists(t *testing.T) {

	//Context
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	user := models.User{
		Name:     "test",
		Email:    "test@test.com",
		Password: "123456",
		Role:     "user",
	}

	mockRepo.On("FindUserByEmail", mock.MatchedBy(func(email string) bool {
		return true
	})).Return(&user, nil)

	_, err := RegisterUser(ctx, mockRepo, user.Name, user.Email, user.Password, user.Role)

	assert.ErrorContains(t, err, "email already registered")
}

func TestGetUserCredential_Success(t *testing.T) {

	//Context
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	user := models.User{
		Name:     "test",
		Email:    "test@test.com",
		Password: "123456",
		Role:     "user",
	}

	mockRepo.On("FindUserByEmail", mock.MatchedBy(func(email string) bool {
		return true
	})).Return(&user, nil)

	result, err := GetUserCredential(ctx, mockRepo, user.Email)
	assert.NoError(t, err)
	assert.Equal(t, result.Email, user.Email)
	assert.Equal(t, result.Name, user.Name)
	assert.Equal(t, result.Role, user.Role)
	assert.NotEmpty(t, result.Password)
}

func TestGetUserCredential_UserNotFound(t *testing.T) {
	//Context
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	user := models.User{
		Email: "test@test.com",
	}
	mockRepo.On("FindUserByEmail", mock.Anything).Return(nil, mongo.ErrNoDocuments)
	result, err := GetUserCredential(ctx, mockRepo, user.Email)

	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Nil(t, result)
}

func TestGetUserByID_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	id := "507f1f77bcf86cd799439011"
	oid, err := primitive.ObjectIDFromHex(id)
	assert.NoError(t, err)

	user := models.User{
		ID:    oid,
		Name:  "test",
		Email: "test@test.com",
		Role:  "user",
	}

	mockRepo.On("FindUserByID", mock.Anything, mock.Anything).Return(&user, nil)

	result, err := GetUserByID(ctx, mockRepo, id)

	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, result.Name, user.Name)
	assert.Equal(t, result.Email, user.Email)
	assert.Equal(t, result.ID.Hex(), id)
	assert.Equal(t, result.Role, user.Role)
}

func TestGetUserByID_UserNotFound(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	id := "507f1f77bcf86cd799439011"

	mockRepo.On("FindUserByID", mock.Anything, mock.Anything).Return(nil, mongo.ErrNoDocuments)

	result, err := GetUserByID(ctx, mockRepo, id)
	mockRepo.AssertExpectations(t)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Nil(t, result)
}
