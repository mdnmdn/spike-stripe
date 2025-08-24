package api

import (
	"net/http"
	"net/http/httptest"
	"pa11y-go-wrapper/internal/analysis"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCompletedHTML(t *testing.T) {
	// Create a new analysis service and add a completed analysis
	service := analysis.NewService(10)
	completedAnalysis := &analysis.Analysis{
		ID:        "test-id",
		URL:       "http://example.com",
		Status:    analysis.StatusCompleted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	service.UpdateResult(completedAnalysis.ID, completedAnalysis.Status, nil)

	// Create a new router
	router := NewRouter(service)

	// Create a new request to the /completed/html endpoint
	req, _ := http.NewRequest("GET", "/api/completed/html", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the response is correct
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "<h1>Completed Analyses</h1>")
}

func TestCompletedPDF(t *testing.T) {
	// Create a new analysis service and add a completed analysis
	service := analysis.NewService(10)
	completedAnalysis := &analysis.Analysis{
		ID:        "test-id",
		URL:       "http://example.com",
		Status:    analysis.StatusCompleted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	service.UpdateResult(completedAnalysis.ID, completedAnalysis.Status, nil)

	// Create a new router
	router := NewRouter(service)

	// Create a new request to the /completed/pdf endpoint
	req, _ := http.NewRequest("GET", "/api/completed/pdf", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the response is correct
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
}
