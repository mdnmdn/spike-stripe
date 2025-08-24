package api

import (
	"embed"
	"io/fs"
	"net/http"
	"pa11y-go-wrapper/internal/analysis"

	"github.com/gin-gonic/gin"
)

// NewRouter creates a new Gin router.
func NewRouter(service *analysis.Service, frontendAssets embed.FS) *gin.Engine {
	r := gin.Default()
	h := NewHandlers(service)

	// Serve the frontend
	staticFiles, _ := fs.Sub(frontendAssets, "frontend")
	r.StaticFS("/", http.FS(staticFiles))

	api := r.Group("/api")
	{
		api.POST("/analyze", h.AnalyzeURL)
		api.POST("/queue", h.QueueURL)
		api.GET("/queue", h.GetQueue)
		api.GET("/queue/:id", h.GetQueueItem)
	}

	return r
}
