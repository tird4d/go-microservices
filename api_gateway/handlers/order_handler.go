package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type OrderGatewayHandler struct {
	OrderServiceURL string
}

func NewOrderGatewayHandler() *OrderGatewayHandler {
	url := os.Getenv("ORDER_SERVICE_URL")
	if url == "" {
		url = "http://order-service:8082"
	}
	return &OrderGatewayHandler{OrderServiceURL: url}
}

func (h *OrderGatewayHandler) CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost,
		h.OrderServiceURL+"/api/v1/orders", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.(string))
	req.Header.Set("X-User-Email", email.(string))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "order-service unavailable"})
		return
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.Status(resp.StatusCode)
		return
	}
	c.JSON(resp.StatusCode, result)
}

func (h *OrderGatewayHandler) ListOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet,
		h.OrderServiceURL+"/api/v1/orders", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}
	req.Header.Set("X-User-ID", userID.(string))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "order-service unavailable"})
		return
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.Status(resp.StatusCode)
		return
	}
	c.JSON(resp.StatusCode, result)
}
