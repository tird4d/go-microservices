package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tird4d/go-microservices/order_service/handlers"
	"github.com/tird4d/go-microservices/order_service/models"
	productpb "github.com/tird4d/go-microservices/product_service/proto"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- test doubles ---

type mockProductClient struct {
	products map[string]*productpb.Product
	err      error
}

func (m *mockProductClient) GetProduct(id string) (*productpb.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.products[id]
	if !ok {
		return nil, errors.New("product not found")
	}
	return p, nil
}

type mockStore struct {
	saveErr      error
	savedOrders  []models.Order
	getOrders    []models.Order
	getErr       error
}

func (m *mockStore) SaveOrder(order models.Order) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedOrders = append(m.savedOrders, order)
	return nil
}

func (m *mockStore) GetOrdersByUser(userID string) ([]models.Order, error) {
	return m.getOrders, m.getErr
}

// --- helpers ---

func newRouter(h *handlers.OrderHandler) *gin.Engine {
	r := gin.New()
	r.POST("/api/v1/orders", h.CreateOrder)
	r.GET("/api/v1/orders", h.ListOrders)
	return r
}

func newHandler(pc handlers.ProductGetter, st handlers.OrderStore) *handlers.OrderHandler {
	return &handlers.OrderHandler{
		ProductClient: pc,
		Publisher:     nil, // nil publisher is safe (guarded in handler)
		Store:         st,
	}
}

// --- CreateOrder tests ---

func TestCreateOrder_MissingUserID_Returns401(t *testing.T) {
	h := newHandler(&mockProductClient{}, &mockStore{})
	r := newRouter(h)

	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{{"product_id": "p1", "quantity": 1}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateOrder_InvalidJSON_Returns400(t *testing.T) {
	h := newHandler(&mockProductClient{}, &mockStore{})
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader([]byte(`{broken`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOrder_EmptyItems_Returns400(t *testing.T) {
	h := newHandler(&mockProductClient{}, &mockStore{})
	r := newRouter(h)

	body, _ := json.Marshal(map[string]any{"items": []any{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOrder_ProductNotFound_Returns400(t *testing.T) {
	pc := &mockProductClient{err: errors.New("rpc: not found")}
	h := newHandler(pc, &mockStore{})
	r := newRouter(h)

	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{{"product_id": "missing-product", "quantity": 1}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOrder_InsufficientStock_Returns400(t *testing.T) {
	pc := &mockProductClient{
		products: map[string]*productpb.Product{
			"p1": {Id: "p1", Name: "Widget", Price: 9.99, Stock: 1},
		},
	}
	h := newHandler(pc, &mockStore{})
	r := newRouter(h)

	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{{"product_id": "p1", "quantity": 5}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOrder_Success_Returns201(t *testing.T) {
	pc := &mockProductClient{
		products: map[string]*productpb.Product{
			"p1": {Id: "p1", Name: "Widget", Price: 9.99, Stock: 10},
		},
	}
	store := &mockStore{}
	h := newHandler(pc, store)
	r := newRouter(h)

	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{{"product_id": "p1", "quantity": 2}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-1")
	req.Header.Set("X-User-Email", "user@example.com")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Len(t, store.savedOrders, 1)
	assert.Equal(t, "user-1", store.savedOrders[0].UserID)
	assert.InDelta(t, 19.98, store.savedOrders[0].TotalPrice, 0.001)
}

func TestCreateOrder_StorageError_Returns500(t *testing.T) {
	pc := &mockProductClient{
		products: map[string]*productpb.Product{
			"p1": {Id: "p1", Name: "Widget", Price: 9.99, Stock: 10},
		},
	}
	h := newHandler(pc, &mockStore{saveErr: errors.New("db error")})
	r := newRouter(h)

	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{{"product_id": "p1", "quantity": 1}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- ListOrders tests ---

func TestListOrders_MissingUserID_Returns401(t *testing.T) {
	h := newHandler(&mockProductClient{}, &mockStore{})
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListOrders_ReturnsEmptyArrayNotNull(t *testing.T) {
	h := newHandler(&mockProductClient{}, &mockStore{getOrders: nil})
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[]`, w.Body.String())
}

func TestListOrders_ReturnsOrders(t *testing.T) {
	existing := []models.Order{
		{ID: "ord-1", UserID: "user-1", Status: "confirmed", TotalPrice: 9.99},
	}
	h := newHandler(&mockProductClient{}, &mockStore{getOrders: existing})
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got []models.Order
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	assert.Len(t, got, 1)
	assert.Equal(t, "ord-1", got[0].ID)
}
