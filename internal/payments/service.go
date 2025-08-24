package payments

import (
	"errors"
	"time"

	"github.com/google/uuid"
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
// Either PriceID should be set for predefined prices, or Amount and Currency for ad-hoc payments.
type CheckoutSessionParams struct {
	PriceID  string `json:"priceId,omitempty"`
	Amount   int64  `json:"amount,omitempty"`
	Currency string `json:"currency,omitempty"`
	// Additional optional fields can be added later (mode, success_url, cancel_url, metadata...)
}

// CheckoutSession represents a simplified session response.
type CheckoutSession struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreateCheckoutSession is a placeholder implementation that mocks session creation.
// Replace with stripe-go integration in the spike.
func (s *Service) CreateCheckoutSession(p CheckoutSessionParams) (*CheckoutSession, error) {
	if p.PriceID == "" && (p.Amount <= 0 || p.Currency == "") {
		return nil, errors.New("either priceId or positive amount and currency is required")
	}
	return &CheckoutSession{
		ID:        "sess_" + uuid.New().String(),
		CreatedAt: time.Now(),
	}, nil
}
