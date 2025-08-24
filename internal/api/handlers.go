package api

import (
	"net/http"
	"stripe-go-spike/internal/payments"

	"github.com/gin-gonic/gin"
)

// Handlers holds the dependencies for the API handlers.
type Handlers struct {
	service *payments.Service
}

// NewHandlers creates new handlers.
func NewHandlers(service *payments.Service) *Handlers {
	return &Handlers{service: service}
}

// Health endpoint for basic readiness.
func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Checkout session creation

type CheckoutSessionRequest struct {
	PriceID  string `json:"priceId,omitempty"`
	Amount   int64  `json:"amount,omitempty"`
	Currency string `json:"currency,omitempty"`
}

func (h *Handlers) CreateCheckoutSession(c *gin.Context) {
	var req CheckoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sess, err := h.service.CreateCheckoutSession(payments.CheckoutSessionParams{
		PriceID:  req.PriceID,
		Amount:   req.Amount,
		Currency: req.Currency,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sess)
}

// Webhook receiver (mock). In real spike, verify signature and parse event types.
func (h *Handlers) Webhook(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
