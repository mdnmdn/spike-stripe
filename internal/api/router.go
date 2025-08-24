package api

import (
	"pa11y-go-wrapper/internal/analysis"

	"github.com/gin-gonic/gin"
)

// NewRouter creates a new Gin router.
func NewRouter(service *analysis.Service) *gin.Engine {
	r := gin.Default()
	h := NewHandlers(service)

	api := r.Group("/api")
	{
		api.POST("/analyze", h.AnalyzeURL)
		api.POST("/queue", h.QueueURL)
		api.GET("/queue", h.GetQueue)
		api.GET("/queue/:id", h.GetQueueItem)
		api.GET("/completed/html", h.GetCompletedAnalysesHTML)
		api.GET("/completed/pdf", h.GetCompletedAnalysesPDF)
	}

	return r
}
