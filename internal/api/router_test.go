package api

import (
	"bytes"
	"embed"
	"net/http"
	"net/http/httptest"
	"stripe-go-spike/internal/payments"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Use an empty embed.FS for tests; router handles missing assets gracefully.
var frontendAssets embed.FS

func TestHealth(t *testing.T) {
	service := payments.NewService(payments.Config{})
	router := NewRouter(service, frontendAssets)

	req, _ := http.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"ok\"")
}

func TestCreateCheckoutSession(t *testing.T) {
	service := payments.NewService(payments.Config{})
	router := NewRouter(service, frontendAssets)

	reqBody := `{"amount": 1000, "currency": "usd"}`
	req := httptest.NewRequest("POST", "/api/checkout-session", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"id\"")
}
