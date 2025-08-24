package api

import (
	"database/sql"
	"io"
	"net/http"
	"strconv"
	"stripe-go-spike/internal/audit"
	"stripe-go-spike/internal/data"
	"stripe-go-spike/internal/db"
	"stripe-go-spike/internal/payments"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handlers holds the dependencies for the API handlers.
type Handlers struct {
	service      *payments.Service
	db           *sql.DB
	queries      *db.Queries
	auditService *audit.Service
}

// NewHandlers creates new handlers.
func NewHandlers(service *payments.Service, database *sql.DB, queries *db.Queries, auditService *audit.Service) *Handlers {
	return &Handlers{
		service:      service,
		db:           database,
		queries:      queries,
		auditService: auditService,
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

type AuditEventsResponse struct {
	Events []data.AuditEvent `json:"events"`
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

	// Log transaction creation start
	h.auditService.LogPayment(c.Request.Context(), "transaction.created",
		"Transaction created for checkout session",
		&req.UserID,
		map[string]interface{}{
			"transaction_id": transactionID,
			"user_id":        req.UserID,
			"product_id":     req.ProductID,
			"product_name":   product.Name,
			"amount":         product.Price,
			"currency":       "usd",
		})

	// Create Stripe checkout session
	sess, err := h.service.CreateCheckoutSession(payments.CheckoutSessionParams{
		Amount:        product.Price,
		Currency:      "usd",
		UserID:        req.UserID,
		ProductID:     req.ProductID,
		TransactionID: transactionID,
	})
	if err != nil {
		// Log checkout session creation failure
		h.auditService.LogStripe(c.Request.Context(), "checkout_session.failed",
			"Failed to create Stripe checkout session",
			&req.UserID,
			map[string]interface{}{
				"transaction_id": transactionID,
				"error":          err.Error(),
			})
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
		h.auditService.LogStripe(c.Request.Context(), "webhook.read_failed",
			"Failed to read webhook request body",
			nil,
			map[string]interface{}{
				"error": err.Error(),
			})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Get the Stripe signature header
	signature := c.GetHeader("Stripe-Signature")

	// Log webhook received
	h.auditService.LogStripe(c.Request.Context(), "webhook.received",
		"Stripe webhook event received",
		nil,
		map[string]interface{}{
			"body_length":   len(body),
			"has_signature": signature != "",
			"signature":     signature,
			"raw_body":      string(body),
		})

	// Process the webhook event
	event, err := h.service.ProcessWebhook(body, signature)
	if err != nil {
		// Log webhook processing failure
		h.auditService.LogStripe(c.Request.Context(), "webhook.processing_failed",
			"Failed to process webhook event",
			nil,
			map[string]interface{}{
				"error":     err.Error(),
				"signature": signature,
			})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log successful webhook processing
	h.auditService.LogStripe(c.Request.Context(), "webhook.processed",
		"Stripe webhook event processed successfully",
		nil,
		map[string]interface{}{
			"event_type": event.Type,
			"session_id": event.SessionID,
			"status":     event.Status,
		})

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		if event.SessionID != "" {
			// Log session completion
			h.auditService.LogStripe(c.Request.Context(), "checkout_session.completed",
				"Checkout session completed",
				nil,
				map[string]interface{}{
					"session_id": event.SessionID,
				})

			// Update transaction status in database
			now := time.Now().UTC().Format(time.RFC3339)
			err = h.queries.UpdateTransactionWithStripeData(c.Request.Context(), db.UpdateTransactionWithStripeDataParams{
				StripeSessionID:       sql.NullString{String: event.SessionID, Valid: true},
				StripePaymentIntentID: sql.NullString{String: "", Valid: false}, // Will be updated by payment_intent.succeeded
				Status:                "completed",
				UpdatedAt:             now,
			})
			if err != nil {
				// Log database update failure
				h.auditService.LogPayment(c.Request.Context(), "transaction.update_failed",
					"Failed to update transaction status",
					nil,
					map[string]interface{}{
						"session_id": event.SessionID,
						"error":      err.Error(),
					})
				// Log error but don't fail the webhook
				// In production, you might want to queue this for retry
			} else {
				// Log successful transaction update
				h.auditService.LogPayment(c.Request.Context(), "transaction.completed",
					"Transaction marked as completed",
					nil,
					map[string]interface{}{
						"session_id": event.SessionID,
					})
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

// GetAuditEvents returns audit events with optional filtering
func (h *Handlers) GetAuditEvents(c *gin.Context) {
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

	subsystem := c.Query("subsystem")
	eventType := c.Query("event_type")
	userID := c.Query("user_id")

	var events []db.AuditEvent
	var err error

	// Query based on filters
	if subsystem != "" && eventType != "" {
		events, err = h.queries.GetAuditEventsBySubsystemAndType(c.Request.Context(), db.GetAuditEventsBySubsystemAndTypeParams{
			Subsystem: subsystem,
			EventType: eventType,
			Limit:     limit,
			Offset:    offset,
		})
	} else if subsystem != "" {
		events, err = h.queries.GetAuditEventsBySubsystem(c.Request.Context(), db.GetAuditEventsBySubsystemParams{
			Subsystem: subsystem,
			Limit:     limit,
			Offset:    offset,
		})
	} else if eventType != "" {
		events, err = h.queries.GetAuditEventsByEventType(c.Request.Context(), db.GetAuditEventsByEventTypeParams{
			EventType: eventType,
			Limit:     limit,
			Offset:    offset,
		})
	} else if userID != "" {
		events, err = h.queries.GetAuditEventsByUser(c.Request.Context(), db.GetAuditEventsByUserParams{
			UserID: sql.NullString{String: userID, Valid: true},
			Limit:  limit,
			Offset: offset,
		})
	} else {
		events, err = h.queries.GetAllAuditEvents(c.Request.Context(), db.GetAllAuditEventsParams{
			Limit:  limit,
			Offset: offset,
		})
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit events"})
		return
	}

	// Convert to API response format
	auditEvents := make([]data.AuditEvent, len(events))
	for i, event := range events {
		timestamp, _ := time.Parse("2006-01-02 15:04:05", event.Timestamp)
		auditEvents[i] = data.AuditEvent{
			ID:        event.ID,
			Timestamp: timestamp,
			Subsystem: event.Subsystem,
			EventType: event.EventType,
			UserID: func() *string {
				if event.UserID.Valid {
					return &event.UserID.String
				}
				return nil
			}(),
			Information: func() *string {
				if event.Information.Valid {
					return &event.Information.String
				}
				return nil
			}(),
			Payload: func() *string {
				if event.Payload.Valid {
					return &event.Payload.String
				}
				return nil
			}(),
		}
	}

	c.JSON(http.StatusOK, AuditEventsResponse{Events: auditEvents})
}
