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

	// ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	// defer cancel()

	userIDRaw, exists := c.Get("user_id")
	userID, ok := userIDRaw.(string)

	if !exists || !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user ID"})
		c.Abort()
		return
	}

	email, exists := c.Get("email")

	if !exists || !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":   email,
		"user_id": userID,
		"message": "this is user profile",
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
