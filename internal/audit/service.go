package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"stripe-go-spike/internal/db"
)

// Service handles audit event logging
type Service struct {
	queries *db.Queries
}

// NewService creates a new audit service
func NewService(queries *db.Queries) *Service {
	return &Service{
		queries: queries,
	}
}

// Event represents an audit event
type Event struct {
	Subsystem   string
	EventType   string
	UserID      *string
	Information string
	Payload     interface{}
	RefID       *string // Primary reference ID (e.g., payment_intent_id)
	RefID2      *string // Secondary reference ID (e.g., session_id)
}

// Log records an audit event
func (s *Service) Log(ctx context.Context, event Event) error {
	var payloadJSON sql.NullString

	if event.Payload != nil {
		payloadBytes, err := json.Marshal(event.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		payloadJSON = sql.NullString{String: string(payloadBytes), Valid: true}
	}

	var userID sql.NullString
	if event.UserID != nil {
		userID = sql.NullString{String: *event.UserID, Valid: true}
	}

	var refID sql.NullString
	if event.RefID != nil {
		refID = sql.NullString{String: *event.RefID, Valid: true}
	}

	var refID2 sql.NullString
	if event.RefID2 != nil {
		refID2 = sql.NullString{String: *event.RefID2, Valid: true}
	}

	return s.queries.CreateAuditEvent(ctx, db.CreateAuditEventParams{
		Subsystem:   event.Subsystem,
		EventType:   event.EventType,
		UserID:      userID,
		Information: sql.NullString{String: event.Information, Valid: event.Information != ""},
		Payload:     payloadJSON,
		RefID:       refID,
		RefId2:      refID2,
	})
}

// LogStripe logs a Stripe-related event
func (s *Service) LogStripe(ctx context.Context, eventType, information string, userID *string, payload interface{}) error {
	return s.Log(ctx, Event{
		Subsystem:   "stripe",
		EventType:   eventType,
		UserID:      userID,
		Information: information,
		Payload:     payload,
	})
}

// LogStripeWithRefs logs a Stripe-related event with reference IDs
func (s *Service) LogStripeWithRefs(ctx context.Context, eventType, information string, userID *string, payload interface{}, refID, refID2 *string) error {
	return s.Log(ctx, Event{
		Subsystem:   "stripe",
		EventType:   eventType,
		UserID:      userID,
		Information: information,
		Payload:     payload,
		RefID:       refID,
		RefID2:      refID2,
	})
}

// LogPayment logs a payment-related event
func (s *Service) LogPayment(ctx context.Context, eventType, information string, userID *string, payload interface{}) error {
	return s.Log(ctx, Event{
		Subsystem:   "payment",
		EventType:   eventType,
		UserID:      userID,
		Information: information,
		Payload:     payload,
	})
}

// LogPaymentWithRefs logs a payment-related event with reference IDs
func (s *Service) LogPaymentWithRefs(ctx context.Context, eventType, information string, userID *string, payload interface{}, refID, refID2 *string) error {
	return s.Log(ctx, Event{
		Subsystem:   "payment",
		EventType:   eventType,
		UserID:      userID,
		Information: information,
		Payload:     payload,
		RefID:       refID,
		RefID2:      refID2,
	})
}

// LogSystem logs a system-related event
func (s *Service) LogSystem(ctx context.Context, eventType, information string, payload interface{}) error {
	return s.Log(ctx, Event{
		Subsystem:   "system",
		EventType:   eventType,
		Information: information,
		Payload:     payload,
	})
}
