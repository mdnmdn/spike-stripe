package api

import (
	"database/sql"
	"embed"
	"io/fs"
	"net/http"
	"stripe-go-spike/internal/audit"
	"stripe-go-spike/internal/db"
	"stripe-go-spike/internal/payments"

	"github.com/gin-gonic/gin"
)

// NewRouter creates a new Gin router.
func NewRouter(service *payments.Service, database *sql.DB, queries *db.Queries, frontendAssets embed.FS) *gin.Engine {
	r := gin.Default()
	auditService := audit.NewService(queries)
	h := NewHandlers(service, database, queries, auditService)

	api := r.Group("/api")
	{
		api.GET("/health", h.Health)
		api.GET("/products", h.GetProducts)
		api.GET("/users", h.GetUsers)
		api.GET("/transactions/:user_id", h.GetUserTransactions)
		api.GET("/transactions", h.GetAllTransactions)
		api.POST("/checkout-session", h.CreateCheckoutSession)
		api.POST("/webhook", h.Webhook)
		api.GET("/audit-events", h.GetAuditEvents)
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
