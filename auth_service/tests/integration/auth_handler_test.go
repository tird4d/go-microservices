package integration

import (
	"context"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang/mock/gomock"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/tird4d/go-microservices/auth_service/config"
	"github.com/tird4d/go-microservices/auth_service/handlers"
	"github.com/tird4d/go-microservices/auth_service/logger"
	"github.com/tird4d/go-microservices/auth_service/mocks"
	authpb "github.com/tird4d/go-microservices/auth_service/proto"
	"github.com/tird4d/go-microservices/auth_service/utils"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env")

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

const bufSize = 1024 * 1024 // 1 MB

var lis *bufconn.Listener

func dialer() func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func startTestGRPCServer(t *testing.T, userClient userpb.UserServiceClient) {
	lis = bufconn.Listen(bufSize)

	server := grpc.NewServer()
	authpb.RegisterAuthServiceServer(server, &handlers.AuthServer{
		UserClient: userClient,
	})

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Fatalf("server exited with error: %v", err)
		}
	}()
}

func TestAuthServer_Login_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)

	// داده‌های تست
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)
	userID := primitive.NewObjectID().Hex()

	mockUserClient.EXPECT().
		GetUserCredential(gomock.Any(), &userpb.GetUserCredentialRequest{Email: email}).
		Return(&userpb.UserCredentialResponse{
			Id:       userID,
			Email:    email,
			Password: hashedPassword,
		}, nil)

	startTestGRPCServer(t, mockUserClient)

	// ساختن کلاینت
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer()), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := authpb.NewAuthServiceClient(conn)

	// ارسال درخواست
	resp, err := client.Login(ctx, &authpb.LoginRequest{
		Email:    email,
		Password: password,
	})

	// بررسی نتیجه
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Login successful", resp.Message)
}

func TestAuthServer_Login_Unsuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)

	// داده‌های تست
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)
	userID := primitive.NewObjectID().Hex()

	mockUserClient.EXPECT().
		GetUserCredential(gomock.Any(), &userpb.GetUserCredentialRequest{Email: email}).
		Return(&userpb.UserCredentialResponse{
			Id:       userID,
			Email:    email,
			Password: hashedPassword,
		}, nil)

	startTestGRPCServer(t, mockUserClient)

	// ساختن کلاینت
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer()), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := authpb.NewAuthServiceClient(conn)

	// ارسال درخواست
	resp, err := client.Login(ctx, &authpb.LoginRequest{
		Email:    email,
		Password: "wrongpassword",
	})

	// بررسی نتیجه
	assert.Empty(t, resp.Token)
	assert.Empty(t, resp.RefreshToken)
	assert.Contains(t, resp.Message, "invalid password")
}

func TestAuthServer_ValidateRefreshToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)

	// داده‌های تست
	userID := primitive.NewObjectID().Hex()
	name := "test"
	email := "test@example.com"

	mockUserClient.EXPECT().
		GetUser(gomock.Any(), &userpb.GetUserRequest{Id: userID}).
		Return(&userpb.UserResponse{
			Id:    userID,
			Name:  name,
			Email: email,
		}, nil)

	startTestGRPCServer(t, mockUserClient)

	ctx := context.Background()

	refreshToken := "valid_refresh_token"

	// Set the refresh token in Redis
	config.RedisClient.Set(ctx, refreshToken, userID, time.Minute)

	// ساختن کلاینت
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer()), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := authpb.NewAuthServiceClient(conn)

	// ارسال درخواست
	resp, err := client.ValidateRefreshToken(ctx, &authpb.ValidateRefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	})

	// بررسی نتیجه
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.NotEqual(t, resp.RefreshToken, refreshToken)

	// Check if the new refresh token is stored in Redis
	redisUserID, err := config.RedisClient.Get(ctx, resp.RefreshToken).Result()
	assert.Equal(t, userID, redisUserID)

	// Check if the old refresh token is deleted from Redis
	_, err = config.RedisClient.Get(ctx, refreshToken).Result()

	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
}

func TestAuthServer_ValidateRefreshToken_TokenNotInRedis(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	mockUserClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Times(0)
	startTestGRPCServer(t, mockUserClient)

	ctx := context.Background()

	// ساختن کلاینت
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer()), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := authpb.NewAuthServiceClient(conn)

	// ارسال درخواست
	resp, err := client.ValidateRefreshToken(ctx, &authpb.ValidateRefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	})

	// بررسی نتیجه
	assert.Empty(t, resp)
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "Invalid refresh token")

}

func TestAuthServer_Logout_Success(t *testing.T) {

	// داده‌های تست
	refreshToken := "valid_refresh_token"
	ctx := context.Background()

	// Set the refresh token in Redis
	config.RedisClient.Set(ctx, refreshToken, "user_id", time.Minute)

	// ساختن کلاینت
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer()), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := authpb.NewAuthServiceClient(conn)

	// ارسال درخواست
	resp, err := client.Logout(ctx, &authpb.LogoutRequest{
		RefreshToken: refreshToken,
	})

	// بررسی نتیجه
	assert.NoError(t, err)
	assert.Contains(t, "Logout successful", resp.Message)

	// Check if the refresh token is deleted from Redis
	_, err = config.RedisClient.Get(ctx, refreshToken).Result()
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
}

func TestAuthServer_Logout_TokenNotInRedis(t *testing.T) {

	ctx := context.Background()

	// ساختن کلاینت
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer()), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := authpb.NewAuthServiceClient(conn)

	// ارسال درخواست
	resp, err := client.Logout(ctx, &authpb.LogoutRequest{
		RefreshToken: "valid_refresh_token",
	})

	// بررسی نتیجه
	assert.Empty(t, resp)
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "invalid refresh token")

}
