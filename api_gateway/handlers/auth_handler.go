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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refresh_token": res.RefreshToken,
		"access_token":  res.AccessToken,
		"message":       "this is user profile",
	})

}

func (h *GatewayHandler) LoginHandler(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.AuthClient.Login(c, &authpb.LoginRequest{
		Email:    body.Email,
		Password: body.Password,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         res.Token,
		"refresh_token": res.RefreshToken,
		"message":       res.Message,
	})
}

func (h *GatewayHandler) LogoutHandler(c *gin.Context) {

	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.AuthClient.Logout(c, &authpb.LogoutRequest{
		RefreshToken: body.RefreshToken,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})

}
