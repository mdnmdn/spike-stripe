package api

import (
	"database/sql"
	"io"
	"net/http"
	"strconv"
	"stripe-go-spike/internal/data"
	"stripe-go-spike/internal/db"
	"stripe-go-spike/internal/payments"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handlers holds the dependencies for the API handlers.
type Handlers struct {
	service *payments.Service
	db      *sql.DB
	queries *db.Queries
}

// NewHandlers creates new handlers.
func NewHandlers(service *payments.Service, database *sql.DB, queries *db.Queries) *Handlers {
	return &Handlers{
		service: service,
		db:      database,
		queries: queries,
	}
}

// Health endpoint for basic readiness.
func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Checkout session creation

type CheckoutSessionRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	ProductID string `json:"product_id" binding:"required"`
}

type CheckoutSessionResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

type ProductsResponse struct {
	Products []data.Product `json:"products"`
}

type UsersResponse struct {
	Users []data.User `json:"users"`
}

type TransactionsResponse struct {
	Transactions []data.Transaction `json:"transactions"`
}

func (h *Handlers) CreateCheckoutSession(c *gin.Context) {
	var req CheckoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate user exists
	user := data.GetUserByID(req.UserID)
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Validate product exists
	product := data.GetProductByID(req.ProductID)
	if product == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Create transaction record
	transactionID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	// Create Stripe checkout session
	sess, err := h.service.CreateCheckoutSession(payments.CheckoutSessionParams{
		Amount:        product.Price,
		Currency:      "usd",
		UserID:        req.UserID,
		ProductID:     req.ProductID,
		TransactionID: transactionID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save transaction to database
	err = h.queries.CreateTransaction(c.Request.Context(), db.CreateTransactionParams{
		ID:              transactionID,
		UserID:          req.UserID,
		ProductID:       req.ProductID,
		ProductName:     product.Name,
		Amount:          product.Price,
		StripeSessionID: sql.NullString{String: sess.ID, Valid: true},
		Status:          "pending",
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	c.JSON(http.StatusOK, CheckoutSessionResponse{
		SessionID: sess.ID,
		URL:       sess.URL,
	})
}

// GetProducts returns the list of hardcoded products
func (h *Handlers) GetProducts(c *gin.Context) {
	c.JSON(http.StatusOK, ProductsResponse{Products: data.Products})
}

// GetUsers returns the list of hardcoded users
func (h *Handlers) GetUsers(c *gin.Context) {
	c.JSON(http.StatusOK, UsersResponse{Users: data.Users})
}

// GetUserTransactions returns transactions for a specific user
func (h *Handlers) GetUserTransactions(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Validate user exists
	user := data.GetUserByID(userID)
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	limit := int64(50)
	offset := int64(0)

	// Parse pagination parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 64); err == nil && o >= 0 {
			offset = o
		}
	}

	txns, err := h.queries.ListTransactionsByUserID(c.Request.Context(), db.ListTransactionsByUserIDParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	transactions := make([]data.Transaction, len(txns))
	for i, txn := range txns {
		createdAt, _ := time.Parse(time.RFC3339, txn.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, txn.UpdatedAt)
		transactions[i] = data.Transaction{
			ID:          txn.ID,
			UserID:      txn.UserID,
			ProductID:   txn.ProductID,
			ProductName: txn.ProductName,
			Amount:      txn.Amount,
			StripeSessionID: func() *string {
				if txn.StripeSessionID.Valid {
					return &txn.StripeSessionID.String
				}
				return nil
			}(),
			StripePaymentIntentID: func() *string {
				if txn.StripePaymentIntentID.Valid {
					return &txn.StripePaymentIntentID.String
				}
				return nil
			}(),
			Status:    txn.Status,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
	}

	c.JSON(http.StatusOK, TransactionsResponse{Transactions: transactions})
}

// GetAllTransactions returns all transactions (admin view)
func (h *Handlers) GetAllTransactions(c *gin.Context) {
	limit := int64(50)
	offset := int64(0)

	// Parse pagination parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 64); err == nil && o >= 0 {
			offset = o
		}
	}

	txns, err := h.queries.ListAllTransactions(c.Request.Context(), db.ListAllTransactionsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	transactions := make([]data.Transaction, len(txns))
	for i, txn := range txns {
		createdAt, _ := time.Parse(time.RFC3339, txn.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, txn.UpdatedAt)
		transactions[i] = data.Transaction{
			ID:          txn.ID,
			UserID:      txn.UserID,
			ProductID:   txn.ProductID,
			ProductName: txn.ProductName,
			Amount:      txn.Amount,
			StripeSessionID: func() *string {
				if txn.StripeSessionID.Valid {
					return &txn.StripeSessionID.String
				}
				return nil
			}(),
			StripePaymentIntentID: func() *string {
				if txn.StripePaymentIntentID.Valid {
					return &txn.StripePaymentIntentID.String
				}
				return nil
			}(),
			Status:    txn.Status,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
	}

	c.JSON(http.StatusOK, TransactionsResponse{Transactions: transactions})
}

// Webhook receiver for Stripe events
func (h *Handlers) Webhook(c *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Get the Stripe signature header
	signature := c.GetHeader("Stripe-Signature")

	// Process the webhook event
	event, err := h.service.ProcessWebhook(body, signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		if event.SessionID != "" {
			// Update transaction status in database
			now := time.Now().UTC().Format(time.RFC3339)
			err = h.queries.UpdateTransactionWithStripeData(c.Request.Context(), db.UpdateTransactionWithStripeDataParams{
				StripeSessionID:       sql.NullString{String: event.SessionID, Valid: true},
				StripePaymentIntentID: sql.NullString{String: "", Valid: false}, // Will be updated by payment_intent.succeeded
				Status:                "completed",
				UpdatedAt:             now,
			})
			if err != nil {
				// Log error but don't fail the webhook
				// In production, you might want to queue this for retry
			}
		}

	case "payment_intent.succeeded":
		// Payment completed successfully
		// Additional processing could be done here

	case "payment_intent.payment_failed":
		// Mark transaction as failed
		if event.SessionID != "" {
			now := time.Now().UTC().Format(time.RFC3339)
			err = h.queries.UpdateTransactionWithStripeData(c.Request.Context(), db.UpdateTransactionWithStripeDataParams{
				StripeSessionID:       sql.NullString{String: event.SessionID, Valid: true},
				StripePaymentIntentID: sql.NullString{String: "", Valid: false},
				Status:                "failed",
				UpdatedAt:             now,
			})
			if err != nil {
				// Log error but don't fail the webhook
			}
		}

	case "checkout.session.expired":
		// Mark transaction as cancelled
		if event.SessionID != "" {
			now := time.Now().UTC().Format(time.RFC3339)
			err = h.queries.UpdateTransactionWithStripeData(c.Request.Context(), db.UpdateTransactionWithStripeDataParams{
				StripeSessionID:       sql.NullString{String: event.SessionID, Valid: true},
				StripePaymentIntentID: sql.NullString{String: "", Valid: false},
				Status:                "cancelled",
				UpdatedAt:             now,
			})
			if err != nil {
				// Log error but don't fail the webhook
			}
		}
	}

	c.Status(http.StatusOK)
}
