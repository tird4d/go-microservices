package services

import (
	"context"
	"errors"
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
	logger.InitLogger(true)

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

func TestGetAllUsers(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	users := []*models.User{
		{
			ID:    primitive.NewObjectID(),
			Name:  "test1",
			Email: "test@test.com",
			Role:  "user",
		},
		{
			ID:    primitive.NewObjectID(),
			Name:  "test2",
			Email: "test2@gmail.com",
			Role:  "admin",
		},
	}
	mockRepo.On("CountUsers", mock.Anything).Return(int64(len(users)), nil)
	mockRepo.On("FindUsers", mock.Anything, mock.Anything, mock.Anything).Return(users, nil)

	page := int32(1)
	pageSize := int32(10)
	result, totalCount, err := GetAllUsers(ctx, mockRepo, int64(page), int64(pageSize))
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "CountUsers", 1)
	mockRepo.AssertNumberOfCalls(t, "FindUsers", 1)
	assert.NoError(t, err)
	assert.Len(t, result, len(users))
	assert.Equal(t, totalCount, int64(len(users)))
	for i, user := range result {
		assert.Equal(t, user.Name, users[i].Name)
		assert.Equal(t, user.Email, users[i].Email)
		assert.Equal(t, user.Role, users[i].Role)
		assert.NotEmpty(t, user.ID)
	}
}

func TestGetAllUsers_EmptyResult(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	mockRepo.On("CountUsers", mock.Anything).Return(int64(0), nil)
	mockRepo.On("FindUsers", mock.Anything, mock.Anything, mock.Anything).Return([]*models.User{}, nil)

	page := int32(1)
	pageSize := int32(10)
	result, totalCount, err := GetAllUsers(ctx, mockRepo, int64(page), int64(pageSize))
	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Equal(t, totalCount, int64(0))
}
func TestGetAllUsers_Error(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	mockRepo.On("FindUsers", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("database error"))

	page := int32(1)
	pageSize := int32(10)
	result, totalCount, err := GetAllUsers(ctx, mockRepo, int64(page), int64(pageSize))
	mockRepo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, totalCount, int64(0))
}

func TestUpdateUser_Success(t *testing.T) {
	//Context
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	id := primitive.NewObjectID()
	user := models.User{
		ID:       id,
		Name:     "test",
		Email:    "test@test.com",
		Password: "123456",
		Role:     "user",
	}
	mockRepo.On("FindUserByEmail", mock.Anything).Return(&user, nil)

	mockRepo.On("UpdateUser", mock.Anything, mock.Anything, mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil)
	result, err := UpdateUser(ctx, mockRepo, id, map[string]any{
		"name":  "updated",
		"email": "updated@test.com",
		"role":  "user",
	})
	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.MatchedCount)
	assert.Equal(t, int64(1), result.ModifiedCount)
}
func TestUpdateUser_UserNotFound(t *testing.T) {
	//Context
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	id := primitive.NewObjectID()

	mockRepo.On("FindUserByEmail", mock.Anything).Return(nil, nil)
	mockRepo.On("UpdateUser", mock.Anything, mock.Anything, mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 0}, nil)

	result, err := UpdateUser(ctx, mockRepo, id, map[string]any{
		"name":  "updated",
		"email": "updated@test.com",
		"role":  "user",
	})
	mockRepo.AssertExpectations(t)
	assert.Equal(t, int64(0), result.MatchedCount)
	assert.NoError(t, err)
}

func TestUpdateUser_EmailAlreadyExists(t *testing.T) {
	//Context
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	oid := primitive.NewObjectID()
	// oid and user id should not match to simulate email already exists scenario
	user := models.User{
		ID:    primitive.NewObjectID(),
		Name:  "test",
		Email: "test@test.com",
		Role:  "user",
	}
	mockRepo.On("FindUserByEmail", mock.Anything).Return(&user, nil)

	result, err := UpdateUser(ctx, mockRepo, oid, map[string]any{
		"name":  "updated",
		"email": "test@test.com",
		"role":  "user",
	})
	mockRepo.AssertExpectations(t)
	assert.Nil(t, result)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.AlreadyExists, st.Code())
}

func TestDeleteUser_Success(t *testing.T) {
	//Context
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	id := primitive.NewObjectID()

	mockRepo.On("DeleteUser", mock.Anything, mock.Anything).Return(&mongo.DeleteResult{DeletedCount: 1}, nil)

	result, err := DeleteUser(ctx, mockRepo, id)
	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.DeletedCount)
}
func TestDeleteUser_UserNotFound(t *testing.T) {
	//Context
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	id := primitive.NewObjectID()

	mockRepo.On("DeleteUser", mock.Anything, mock.Anything).Return(&mongo.DeleteResult{DeletedCount: 0}, nil)

	result, err := DeleteUser(ctx, mockRepo, id)
	mockRepo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.DeletedCount)
}
func TestDeleteUser_Error(t *testing.T) {
	//Context
	ctx := context.Background()

	mockRepo := new(mocks.UserRepositoryMock)
	id := primitive.NewObjectID()

	mockRepo.On("DeleteUser", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))

	result, err := DeleteUser(ctx, mockRepo, id)
	mockRepo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, result)
}
