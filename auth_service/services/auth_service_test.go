package services

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang/mock/gomock"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/tird4d/go-microservices/auth_service/config"
	"github.com/tird4d/go-microservices/auth_service/logger"
	"github.com/tird4d/go-microservices/auth_service/mocks"
	"github.com/tird4d/go-microservices/auth_service/utils"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")

	if err != nil {
		logger.Info("Error loading .env file")
	}

	//Create a redis client with in-memory database
	s, err := miniredis.Run()
	if err != nil {
		log.Fatalf("❌ Failed to start mini redis: %v", err)
	}

	config.RedisClient = redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	os.Exit(m.Run())

}
func TestLoginUser_Success(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	ctx := context.Background()

	//Test data
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)

	// Answer mock

	mockUserClient.EXPECT().
		GetUserCredential(gomock.Any(), &userpb.GetUserCredentialRequest{Email: email}).
		Return(&userpb.UserCredentialResponse{
			Id:       primitive.NewObjectID().Hex(),
			Email:    email,
			Password: hashedPassword,
		}, nil)

	token, refreshToken, err := LoginUser(ctx, mockUserClient, email, password)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, refreshToken)
}

func TestLoginUser_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	ctx := context.Background()

	//Test data
	email := "test@example.com"
	wrongPassword := "wrong123"
	hashedPassword, _ := utils.HashPassword("correct123")

	// Answer mock
	mockUserClient.EXPECT().
		GetUserCredential(gomock.Any(), &userpb.GetUserCredentialRequest{Email: email}).
		Return(&userpb.UserCredentialResponse{
			Id:       primitive.NewObjectID().Hex(),
			Email:    email,
			Password: hashedPassword,
		}, nil)

		// Call the function
	token, refreshToken, err := LoginUser(ctx, mockUserClient, email, wrongPassword)
	// Check the result
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
}

func TestLoginUser_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	ctx := context.Background()

	//Test data
	email := "wrong@example.com"
	password := "password123"

	// Answer mock
	mockUserClient.EXPECT().
		GetUserCredential(gomock.Any(), &userpb.GetUserCredentialRequest{Email: email}).
		Return(&userpb.UserCredentialResponse{
			Id:       "",
			Email:    "",
			Password: "",
		}, status.Error(codes.NotFound, "email not found"))

	// Call the function
	token, refreshToken, err := LoginUser(ctx, mockUserClient, email, password)
	// Check the result
	assert.ErrorIs(t, err, status.Error(codes.NotFound, "email not found"))
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
}

func TestLoginUser_InvalidUserIDFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	ctx := context.Background()

	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)

	// wrong user ID format
	mockUserClient.EXPECT().
		GetUserCredential(gomock.Any(), &userpb.GetUserCredentialRequest{Email: email}).
		Return(&userpb.UserCredentialResponse{
			Id:       "NOT_A_VALID_OBJECT_ID",
			Email:    email,
			Password: hashedPassword,
		}, nil)

	token, refreshToken, err := LoginUser(ctx, mockUserClient, email, password)

	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
}

func TestValidateRefreshToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	ctx := context.Background()

	//Test data
	email := "test@example.com"
	refreshToken := "valid_refresh_token"
	userID := primitive.NewObjectID().Hex()
	// Set the refresh token in Redis
	config.RedisClient.Set(ctx, refreshToken, userID, time.Minute)

	// Answer mock
	mockUserClient.EXPECT().
		GetUser(gomock.Any(), &userpb.GetUserRequest{Id: userID}).
		Return(&userpb.UserResponse{
			Id:    userID,
			Name:  "Test User",
			Email: email,
		}, nil)

	// Call the function
	token, newRefreshToken, err := ValidateRefreshToken(ctx, mockUserClient, refreshToken)
	// Check the result
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, refreshToken, newRefreshToken)
	// Check if the new refresh token is stored in Redis
	redisUserID, err := config.RedisClient.Get(ctx, newRefreshToken).Result()
	assert.NoError(t, err)
	assert.Equal(t, userID, redisUserID)
	// Check if the old refresh token is deleted from Redis
	_, err = config.RedisClient.Get(ctx, refreshToken).Result()
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
}

func TestValidateRefreshToken_InvalidRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	ctx := context.Background()

	//Test data
	refreshToken := "invalid_refresh_token"

	// Call the function
	token, newRefreshToken, err := ValidateRefreshToken(ctx, mockUserClient, refreshToken)
	// Check the result
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Empty(t, token)
	assert.Empty(t, newRefreshToken)
}

func TestValidateRefreshToken_UserServiceUnavailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	ctx := context.Background()

	//Test data
	refreshToken := "valid_refresh_token"
	userID := primitive.NewObjectID().Hex()
	// Set the refresh token in Redis
	config.RedisClient.Set(ctx, refreshToken, userID, time.Minute)

	// Answer mock
	mockUserClient.EXPECT().
		GetUser(gomock.Any(), &userpb.GetUserRequest{Id: userID}).
		Return(nil, status.Error(codes.Unavailable, "service unavailable"))

	// Call the function
	token, newRefreshToken, err := ValidateRefreshToken(ctx, mockUserClient, refreshToken)
	// Check the result
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unavailable, st.Code())
	assert.Empty(t, token)
	assert.Empty(t, newRefreshToken)
}
func TestValidateRefreshToken_InvalidUserIDFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	ctx := context.Background()

	//Test data
	refreshToken := "valid_refresh_token"
	userID := "NOT_A_VALID_OBJECT_ID"
	// Set the refresh token in Redis
	config.RedisClient.Set(ctx, refreshToken, userID, time.Minute)

	// Answer mock
	mockUserClient.EXPECT().
		GetUser(gomock.Any(), &userpb.GetUserRequest{Id: userID}).
		Return(&userpb.UserResponse{
			Id:    userID,
			Name:  "Test User",
			Email: "test@example.com",
		}, nil)

	// Call the function
	token, newRefreshToken, err := ValidateRefreshToken(ctx, mockUserClient, refreshToken)
	// Check the result
	assert.Error(t, err)
	assert.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid user ID"))
	assert.Empty(t, token)
	assert.Empty(t, newRefreshToken)
}

func TestDeleteRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	refreshToken := "valid_refresh_token"
	userID := primitive.NewObjectID().Hex()
	// Set the refresh token in Redis
	config.RedisClient.Set(ctx, refreshToken, userID, time.Minute)
	// Call the function
	err := DeleteRefreshToken(ctx, refreshToken)
	// Check the result
	assert.NoError(t, err)
	// Check if the refresh token is deleted from Redis
	_, err = config.RedisClient.Get(ctx, refreshToken).Result()
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
	// Check if the refresh token is deleted from Redis
	_, err = config.RedisClient.Get(ctx, refreshToken).Result()
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
}
