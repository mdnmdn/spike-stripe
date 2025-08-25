package payments

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/webhook"
)

// Config holds Stripe-related configuration. Values can be empty for local spikes.
type Config struct {
	SecretKey      string
	PublishableKey string
	WebhookSecret  string
}

// Service provides minimal payment-related operations needed for the spike.
// In a real integration, this would wrap the official stripe-go client.
type Service struct {
	cfg Config
}

func NewService(cfg Config) *Service {
	return &Service{cfg: cfg}
}

// CheckoutSessionParams captures basic parameters to create a checkout session.
type CheckoutSessionParams struct {
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	UserID        string `json:"user_id"`
	ProductID     string `json:"product_id"`
	TransactionID string `json:"transaction_id"`
}

// CheckoutSession represents a simplified session response.
type CheckoutSession struct {
	ID              string    `json:"id"`
	URL             string    `json:"url"`
	PaymentIntentID string    `json:"payment_intent_id,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}

// CreateCheckoutSession creates a Stripe checkout session or returns a mock if no API key is set.
func (s *Service) CreateCheckoutSession(p CheckoutSessionParams) (*CheckoutSession, error) {
	if p.Amount <= 0 || p.Currency == "" {
		return nil, errors.New("positive amount and currency are required")
	}

	// If no Stripe secret key is configured, return mock response
	if s.cfg.SecretKey == "" {
		mockID := uuid.New().String()
		return &CheckoutSession{
			ID:              "sess_mock_" + mockID,
			URL:             fmt.Sprintf("https://checkout.stripe.com/pay/sess_mock_%s", mockID),
			PaymentIntentID: "pi_mock_" + mockID,
			CreatedAt:       time.Now(),
		}, nil
	}

	// Configure Stripe API key
	stripe.Key = s.cfg.SecretKey

	// Get the base URL from environment or use default
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8060"
	}

	// Create Stripe checkout session
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(p.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(fmt.Sprintf("Product %s", p.ProductID)),
					},
					UnitAmount: stripe.Int64(p.Amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(fmt.Sprintf("%s/app?success=true&session_id={CHECKOUT_SESSION_ID}", baseURL)),
		CancelURL:  stripe.String(fmt.Sprintf("%s/app?canceled=true", baseURL)),
		Metadata: map[string]string{
			"user_id":        p.UserID,
			"product_id":     p.ProductID,
			"transaction_id": p.TransactionID,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe session: %w", err)
	}

	checkoutSession := &CheckoutSession{
		ID:        sess.ID,
		URL:       sess.URL,
		CreatedAt: time.Unix(sess.Created, 0),
	}

	// Extract payment intent ID if available
	if sess.PaymentIntent != nil {
		checkoutSession.PaymentIntentID = sess.PaymentIntent.ID
	}

	return checkoutSession, nil
}

// WebhookEvent represents a Stripe webhook event
type WebhookEvent struct {
	Type            string                 `json:"type"`
	Data            map[string]interface{} `json:"data"`
	SessionID       string                 `json:"session_id,omitempty"`
	PaymentIntentID string                 `json:"payment_intent_id,omitempty"`
	Status          string                 `json:"status,omitempty"`
}

// ProcessWebhook processes a Stripe webhook event
func (s *Service) ProcessWebhook(payload []byte, signature string) (*WebhookEvent, error) {
	// If no webhook secret is configured, return mock event
	if s.cfg.WebhookSecret == "" {
		mockID := uuid.New().String()
		return &WebhookEvent{
			Type:            "checkout.session.completed",
			Data:            make(map[string]interface{}),
			PaymentIntentID: "pi_mock_" + mockID,
			Status:          "completed",
		}, nil
	}

	// Verify webhook signature with option to ignore API version mismatch
	event, err := webhook.ConstructEventWithOptions(payload, signature, s.cfg.WebhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		return nil, fmt.Errorf("webhook signature verification failed: %w", err)
	}

	// Extract relevant information based on event type
	webhookEvent := &WebhookEvent{
		Type: string(event.Type),
		Data: event.Data.Object,
	}

	switch event.Type {
	case "checkout.session.completed":
		if sessionData, ok := event.Data.Object["id"].(string); ok {
			webhookEvent.SessionID = sessionData
			webhookEvent.Status = "completed"
		}
		// Extract payment intent ID from session
		if paymentIntentData, ok := event.Data.Object["payment_intent"].(string); ok {
			webhookEvent.PaymentIntentID = paymentIntentData
		}
	case "payment_intent.succeeded":
		webhookEvent.Status = "completed"
		if paymentIntentData, ok := event.Data.Object["id"].(string); ok {
			webhookEvent.PaymentIntentID = paymentIntentData
		}
	case "payment_intent.payment_failed":
		webhookEvent.Status = "failed"
		if paymentIntentData, ok := event.Data.Object["id"].(string); ok {
			webhookEvent.PaymentIntentID = paymentIntentData
		}
	case "checkout.session.expired":
		if sessionData, ok := event.Data.Object["id"].(string); ok {
			webhookEvent.SessionID = sessionData
			webhookEvent.Status = "cancelled"
		}
		// Extract payment intent ID from session
		if paymentIntentData, ok := event.Data.Object["payment_intent"].(string); ok {
			webhookEvent.PaymentIntentID = paymentIntentData
		}
	}

	return webhookEvent, nil
}
