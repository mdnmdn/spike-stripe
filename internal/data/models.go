package data

import "time"

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"` // "user" or "admin"
}

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"` // price in cents
}

type Transaction struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"user_id"`
	ProductID             string    `json:"product_id"`
	ProductName           string    `json:"product_name"`
	Amount                int64     `json:"amount"`
	StripeSessionID       *string   `json:"stripe_session_id,omitempty"`
	StripePaymentIntentID *string   `json:"stripe_payment_intent_id,omitempty"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// Hardcoded users for the spike
var Users = []User{
	{ID: "luke", Name: "Luke", Role: "user"},
	{ID: "jinny", Name: "Jinny", Role: "user"},
	{ID: "admin", Name: "ADMIN", Role: "admin"},
}

// Hardcoded products for the spike
var Products = []Product{
	{
		ID:          "lumaweave",
		Name:        "LumaWeave Reactive Threads",
		Description: "Smart textile technology",
		Price:       4999, // $49.99
	},
	{
		ID:          "coffee-pods",
		Name:        "Atmospheric Coffee Pods",
		Description: "Premium coffee experience",
		Price:       2999, // $29.99
	},
	{
		ID:          "echospout",
		Name:        "EchoSprout Memory Plants",
		Description: "Living memory storage",
		Price:       8999, // $89.99
	},
	{
		ID:          "pocketforge",
		Name:        "PocketForge Nano Printer",
		Description: "Miniature 3D printing",
		Price:       19999, // $199.99
	},
}

// GetUserByID returns a user by ID
func GetUserByID(id string) *User {
	for _, user := range Users {
		if user.ID == id {
			return &user
		}
	}
	return nil
}

// GetProductByID returns a product by ID
func GetProductByID(id string) *Product {
	for _, product := range Products {
		if product.ID == id {
			return &product
		}
	}
	return nil
}

// AuditEvent represents an audit event for API responses
type AuditEvent struct {
	ID          int64     `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Subsystem   string    `json:"subsystem"`
	EventType   string    `json:"event_type"`
	UserID      *string   `json:"user_id,omitempty"`
	Information *string   `json:"information,omitempty"`
	Payload     *string   `json:"payload,omitempty"`
}
