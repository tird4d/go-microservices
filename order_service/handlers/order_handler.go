package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tird4d/go-microservices/order_service/clients"
	"github.com/tird4d/go-microservices/order_service/events"
	"github.com/tird4d/go-microservices/order_service/models"
	"github.com/tird4d/go-microservices/order_service/storage"
)

type OrderHandler struct {
	ProductClient *clients.ProductClient
	Publisher     *events.OrderPublisher
}

type CreateOrderRequest struct {
	Items []struct {
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int32  `json:"quantity"   binding:"required,min=1"`
	} `json:"items" binding:"required,min=1"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	userEmail := c.GetHeader("X-User-Email")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing X-User-ID header"})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orderItems []models.OrderItem
	var total float64

	for _, item := range req.Items {
		product, err := h.ProductClient.GetProduct(item.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product not found: " + item.ProductID})
			return
		}
		if product.Stock < item.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient stock for: " + product.Name})
			return
		}
		linePrice := product.Price * float64(item.Quantity)
		total += linePrice
		orderItems = append(orderItems, models.OrderItem{
			ProductID: product.Id,
			Name:      product.Name,
			Price:     product.Price,
			Quantity:  item.Quantity,
		})
	}

	order := models.Order{
		ID:         uuid.New().String(),
		UserID:     userID,
		UserEmail:  userEmail,
		Items:      orderItems,
		TotalPrice: total,
		Status:     "confirmed",
		CreatedAt:  time.Now(),
	}

	if err := storage.SaveOrder(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save order"})
		return
	}

	if h.Publisher != nil {
		h.Publisher.PublishOrderPlaced(order)
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing X-User-ID header"})
		return
	}

	orders, err := storage.GetOrdersByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}

	if orders == nil {
		orders = []models.Order{}
	}
	c.JSON(http.StatusOK, orders)
}
