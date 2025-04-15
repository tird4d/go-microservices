package handlers

import (
	"net/http"

	authpb "github.com/tird4d/go-microservices/auth_service/proto"

	"github.com/gin-gonic/gin"
)

type GatewayHandler struct {
	authpb.UnimplementedAuthServiceServer
	AuthClient authpb.AuthServiceClient
}

func (h *GatewayHandler) RefreshTokenHandler(c *gin.Context) {

	// Get the old refresh token from the request

	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the refresh token and make new one
	res, err := h.AuthClient.ValidateRefreshToken(c, &authpb.ValidateRefreshTokenRequest{
		RefreshToken: body.RefreshToken,
	})

	if err != nil || res == nil {
		println("error in refresh token", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refresh_token": res.RefreshToken,
		"access_token":  res.AccessToken,
		"message":       "this is user profile",
	})

}
