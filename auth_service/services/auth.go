package services

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tird4d/go-microservices/auth_service/config"
	"github.com/tird4d/go-microservices/auth_service/logger"
	"github.com/tird4d/go-microservices/auth_service/utils"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
)

func LoginUser(ctx context.Context, userClient userpb.UserServiceClient, email, password string) (string, string, error) {

	res, err := userClient.GetUserCredential(ctx, &userpb.GetUserCredentialRequest{
		Email: email,
	})
	log.Printf("🔍 User credential response: %v", res)

	if res, err = userServiceResponseHandler(res, err); err != nil {
		return "", "", err
	}

	ok := utils.CheckPasswordHash(password, res.Password)
	if !ok {
		logger.Error("Invalid password: %v", err)
		return "", "", status.Errorf(codes.Unauthenticated, "invalid password")
	}

	oid, err := primitive.ObjectIDFromHex(res.Id)
	if err != nil {
		logger.Error("Invalid user ID: %v", err)
		return "", "", status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	token, err := utils.GenerateJWT(oid, res.Email)
	if err != nil {
		logger.Error("Failed to generate JWT: %v", err)
		return "", "", status.Errorf(codes.Internal, "failed to generate JWT")
	}

	refreshToken, err := createRefreshToken(ctx, res.Id)
	if err != nil || refreshToken == "" {
		logger.Error("Failed to create refresh token: %v", err)
		return "", "", status.Errorf(codes.Internal, "failed to create refresh token")
	}

	// delete the old refresh token if it exists

	log.Printf("✅ JWT generated: %s RefreshToken generated %s", token, refreshToken)
	return token, refreshToken, nil

}

func ValidateRefreshToken(ctx context.Context, userClient userpb.UserServiceClient, refreshToken string) (string, string, error) {
	// Check if the refresh token is valid
	if refreshToken == "" {
		return "", "", status.Errorf(codes.Unauthenticated, "invalid refresh token")
	}

	// Validate the refresh token with Redis
	userID, err := config.RedisClient.Get(ctx, refreshToken).Result()

	if err != nil || userID == "" {
		log.Printf("❌ Failed to validate refresh token: %v", err.Error())
		return "", "", status.Errorf(codes.Unauthenticated, "invalid refresh token")
	}

	user, err := userClient.GetUser(ctx, &userpb.GetUserRequest{
		Id: userID,
	})

	if err != nil || user == nil || user.Id == "" {
		log.Printf("❌ Failed to connect to user_service: %v", err.Error())
		return "", "", status.Errorf(codes.Unavailable, "cannot connect to user service")
	}

	// Convert userID to ObjectID
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Printf("❌ Invalid user ID: %v", err)
		return "", "", status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	// Generate a new access token
	token, err := utils.GenerateJWT(oid, user.Email)
	if err != nil {
		log.Printf("❌ Failed to generate JWT: %v", err)
		return "", "", status.Errorf(codes.Internal, "failed to generate JWT")
	}

	// Create a new refresh token
	newRefreshToken, err := createRefreshToken(ctx, userID)

	if err != nil {
		log.Printf("❌ Failed to store refresh token in Redis: %v", err)
		return "", "", status.Errorf(codes.Internal, "failed to store refresh token")
	}

	if newRefreshToken == "" {
		return "", "", status.Errorf(codes.Internal, "failed to generate refresh token")
	}

	// Delete the old refresh token
	err = DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		log.Printf("❌ Failed to delete old refresh token: %v", err)
		return "", "", status.Errorf(codes.Internal, "failed to delete old refresh token")
	}

	log.Printf("✅ New AccessToken and RefreshToken issued for user: %s", userID)

	return token, newRefreshToken, nil
}

func createRefreshToken(ctx context.Context, userID string) (string, error) {
	// Generate a refresh token with a longer expiration time
	refreshToken := utils.GenerateRefreshToken()

	// Store the refresh token in Redis with userID as the key
	err := config.RedisClient.Set(ctx, refreshToken, userID, 7*24*time.Hour).Err()

	return refreshToken, err
}

func DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	_, err := config.RedisClient.Del(ctx, refreshToken).Result()

	return err

}

func userServiceResponseHandler(res *userpb.UserCredentialResponse, err error) (*userpb.UserCredentialResponse, error) {
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		logger.Error("Failed to connect to user_service: %v", err)
		return nil, status.Error(codes.Unavailable, "cannot connect to user service")
	}

	if res == nil {
		logger.Error("userService responded with nil")
		return nil, status.Errorf(codes.Internal, "empty response from user service")
	}

	return res, nil
}
