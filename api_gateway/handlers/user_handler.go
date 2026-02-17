package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
)

type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	UserClient userpb.UserServiceClient
}

func (u *UserHandler) MeHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userIDRaw, exists := c.Get("user_id")
	userID, ok := userIDRaw.(string)

	if !exists || !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user ID"})
		c.Abort()
		return
	}

	// Fetch complete user data from user service
	userRes, err := u.UserClient.GetUser(ctx, &userpb.GetUserRequest{
		Id: userID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data", "details": err.Error()})
		return
	}

	// Return user data in the format expected by frontend
	c.JSON(http.StatusOK, gin.H{
		"id":         userRes.Id,
		"email":      userRes.Email,
		"username":   userRes.Name,     // Map backend 'name' to frontend 'username'
		"name":       userRes.Name,     // Also provide 'name' field
		"role":       userRes.Role,
		"created_at": time.Now().Format(time.RFC3339), // Placeholder
		"updated_at": time.Now().Format(time.RFC3339), // Placeholder
	})
}

func (u *UserHandler) RegisterHandler(c *gin.Context) {
	var body struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role" binding:"required,oneof=admin user"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := u.UserClient.Register(ctx, &userpb.RegisterRequest{
		Name:     body.Name,
		Email:    body.Email,
		Password: body.Password,
		Role:     body.Role,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": res.Id,
		"message": res.Message,
	})

}
