// user_handlers.go
package handlers

import (
	"context"
	"log"

	authpb "github.com/tird4d/go-microservices/auth_service/proto"
	"github.com/tird4d/go-microservices/auth_service/services"
	"github.com/tird4d/go-microservices/auth_service/utils"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	UserClient userpb.UserServiceClient
}

func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	log.Printf("📥 Login called for email: %s and pass is: %s", req.Email, req.Password)

	token, refreshToken, err := services.LoginUser(ctx, s.UserClient, req.Email, req.Password)

	message := "Login successful"
	if err != nil {
		log.Printf("❌ Login failed: %v", err)
		message = err.Error()
	}

	return &authpb.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Message:      message,
	}, nil
}

func (s *AuthServer) Validate(ctx context.Context, req *authpb.ValidateRequest) (*authpb.ValidateResponse, error) {
	log.Printf("🔐 Validate called with token: %s", req.Token)

	claims, err := utils.ValidateJWT(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid or expired token")
	}

	// گرفتن اطلاعات از claims
	userID, ok1 := claims["user_id"].(string)
	email, ok2 := claims["email"].(string)
	role, ok3 := claims["role"].(string)

	if !ok1 || !ok2 || !ok3 {
		return nil, status.Error(codes.Internal, "Invalid token payload")
	}

	return &authpb.ValidateResponse{
		UserId: userID,
		Email:  email,
		Role:   role,
	}, nil
}

func (s *AuthServer) ValidateRefreshToken(ctx context.Context, req *authpb.ValidateRefreshTokenRequest) (*authpb.ValidateRefreshTokenResponse, error) {

	accessToken, RefreshToken, err := services.ValidateRefreshToken(ctx, s.UserClient, req.RefreshToken)
	if err != nil {
		log.Printf("❌ Refresh token validation failed: %v", err)
		return nil, status.Error(codes.Unauthenticated, "Invalid refresh token")
	}

	return &authpb.ValidateRefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: RefreshToken,
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	log.Printf("🔐 Logout called with token: %s", req.RefreshToken)

	err := services.DeleteRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		log.Printf("❌ Logout failed: %v", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &authpb.LogoutResponse{
		Message: "Logout successful",
	}, nil
}
