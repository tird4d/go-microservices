package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
)

type AdminHandler struct {
	userpb.UnimplementedUserServiceServer
	UserClient userpb.UserServiceClient
}

func (a *AdminHandler) UsersHandler(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "all users",
	})
}
