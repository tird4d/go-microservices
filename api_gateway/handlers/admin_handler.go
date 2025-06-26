package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

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

	updateRequest := userpb.UpdateUserRequest{Id: userId}

	if _, ok := c.GetPostForm("name"); ok {
		updateRequest.Name = &wrapperspb.StringValue{Value: body.Name}
	}
	if _, ok := c.GetPostForm("email"); ok {
		updateRequest.Email = &wrapperspb.StringValue{Value: body.Email}
	}
	if _, ok := c.GetPostForm("role"); ok {
		updateRequest.Role = &wrapperspb.StringValue{Value: body.Role}
	}

	updateRequest.Id = userId
	_, err := a.UserClient.UpdateUser(ctx, &updateRequest)

	if err != nil {
		log.Printf("Error updating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"status":  "success",
	})
}

func (a *AdminHandler) DeleteHandler(c *gin.Context) {
	userId := c.Param("user_id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := a.UserClient.DeleteUser(ctx, &userpb.DeleteUserRequest{Id: userId})

	if err != nil {
		if status.Code(err) == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": res.Message,
		"status":  "success",
	})
}
