package api

import (
	"bytes"
	"embed"
	"net/http"
	"net/http/httptest"
	"stripe-go-spike/internal/db"
	"stripe-go-spike/internal/payments"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// Use an empty embed.FS for tests; router handles missing assets gracefully.
var frontendAssets embed.FS

func TestHealth(t *testing.T) {
	service := payments.NewService(payments.Config{})
	// Create in-memory database for testing
	database, err := db.NewTestConnection()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()
	queries := db.New(database)
	router := NewRouter(service, database, queries, frontendAssets)

	req, _ := http.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"ok\"")
}

func TestCreateCheckoutSession(t *testing.T) {
	service := payments.NewService(payments.Config{})
	// Create in-memory database for testing
	database, err := db.NewTestConnection()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()
	queries := db.New(database)
	router := NewRouter(service, database, queries, frontendAssets)

	reqBody := `{"user_id": "luke", "product_id": "lumaweave"}`
	req := httptest.NewRequest("POST", "/api/checkout-session", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"session_id\"")
	assert.Contains(t, w.Body.String(), "\"url\"")
}
