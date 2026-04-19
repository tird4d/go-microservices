package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tird4d/go-microservices/email_service/metrics"
)

func TestMain(m *testing.M) {
	metrics.InitMetrics()
	m.Run()
}

func TestHealthHandler_Returns200(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHealthHandler_ReturnsJSONBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

func TestProcessUserEvent_ValidPayload(t *testing.T) {
	payload := []byte(`{"user_id":"u1","email":"user@example.com","name":"Alice"}`)
	if err := processUserEvent(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProcessUserEvent_InvalidJSON(t *testing.T) {
	if err := processUserEvent([]byte(`not-json`)); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestProcessOrderEvent_ValidPayload(t *testing.T) {
	payload := []byte(`{
		"order_id":"ord1",
		"user_id":"u1",
		"user_email":"user@example.com",
		"items":[{"product_id":"p1","name":"Widget","price":9.99,"quantity":2}],
		"total_price":19.98
	}`)
	if err := processOrderEvent(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProcessOrderEvent_InvalidJSON(t *testing.T) {
	if err := processOrderEvent([]byte(`{broken`)); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
