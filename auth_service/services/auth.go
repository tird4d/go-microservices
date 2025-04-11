package services

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tird4d/go-microservices/auth_service/utils"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
)

func LoginUser(ctx context.Context, userClient userpb.UserServiceClient, email, password string) (string, error) {

	res, err := userClient.GetUserCredential(ctx, &userpb.GetUserCredentialRequest{
		Email: email,
	})
	log.Printf("🔍 User credential response: %v", res)

	if err != nil {
		log.Printf("❌ Failed to connect to user_service: %v", err)
		return "", status.Errorf(codes.Unavailable, "cannot connect to user service")
	}

	if res == nil {
		log.Println("⚠️ userService responded with nil")
		return "", status.Errorf(codes.Internal, "empty response from user service")
	}

	ok := utils.CheckPasswordHash(password, res.Password)
	if !ok {
		log.Println("❌ Invalid password")
		return "", status.Errorf(codes.Unauthenticated, "invalid password")
	}

	oid, err := primitive.ObjectIDFromHex(res.Id)
	if err != nil {
		log.Printf("❌ Invalid user ID: %v", err)
		return "", status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	token, err := utils.GenerateJWT(oid, res.Email)
	if err != nil {
		log.Printf("❌ Failed to generate JWT: %v", err)
		return "", status.Errorf(codes.Internal, "failed to generate JWT")
	}

	log.Printf("✅ JWT generated: %s", token)
	return token, nil

}
