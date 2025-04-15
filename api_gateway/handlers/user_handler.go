package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func MeHandler(c *gin.Context) {

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
