package api

import (
	"embed"
	"io/fs"
	"net/http"
	"stripe-go-spike/internal/payments"

	"github.com/gin-gonic/gin"
)

// NewRouter creates a new Gin router.
func NewRouter(service *payments.Service, frontendAssets embed.FS) *gin.Engine {
	r := gin.Default()
	h := NewHandlers(service) // service is payments.Service

	api := r.Group("/api")
	{
		api.GET("/health", h.Health)
		api.POST("/checkout-session", h.CreateCheckoutSession)
		api.POST("/webhook", h.Webhook)
	}

	// Serve the frontend if available
	if staticFiles, err := fs.Sub(frontendAssets, "frontend"); err == nil {
		r.StaticFS("/app", http.FS(staticFiles))
		r.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/app")
		})
	} else {
		r.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"name": "stripe-go-spike"})
		})
	}

	return r
}
