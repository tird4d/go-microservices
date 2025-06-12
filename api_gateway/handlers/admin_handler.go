package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type AdminHandler struct {
	userpb.UnimplementedUserServiceServer
	UserClient userpb.UserServiceClient
}

func (a *AdminHandler) UsersHandler(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	users, err := a.UserClient.GetAllUsers(ctx, &userpb.GetAllUsersRequest{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to retrieve users",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":   users.Users,
		"message": "Users retrieved successfully",
		"status":  "success",
	})
}

func (a *AdminHandler) UpdateUserHandler(c *gin.Context) {
	userId := c.Param("user_id")
	var body struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateRequest := userpb.UpdateUserRequest{
		Id:    userId,
		Name:  &wrapperspb.StringValue{Value: body.Name},
		Email: &wrapperspb.StringValue{Value: body.Email},
		Role:  &wrapperspb.StringValue{Value: body.Role},
	}
	updateRequest.Id = userId
	res, err := a.UserClient.UpdateUser(ctx, &updateRequest)

	if err != nil {
		log.Printf("Error updating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":    res.Id,
		"message": "User updated successfully",
	})
}
