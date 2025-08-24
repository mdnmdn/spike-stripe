# Description of spike structure 

## UI
- first page a user selector: Luke, Jinny, -ADMIN-, the first two are simple user for testing, the last is an admin user
  - manage the current user with a session storage
- selecting a simple user will show a page with two tabs:
  - buy with a list of 4 products (page product)
  - list of transactions of current user (user transactions)
- selecting an admin user will show a page with a list of transactions of all users (admin transactions)

### Page product
- list of products (LumaWeave Reactive Threads, Atmospheric Coffee Pods, EchoSprout Memory Plants, PocketForge Nano printer) that could be bought with an invented price
- the user could select a product and a CTA will appear: "buy stripe hosted"
- when the user click on the CTA, the stripe hosted payment page will start

## Technical Implementation Requirements

### Hardcoded Data

#### Users (stored in frontend/backend code)
```go
type User struct {
    ID   string
    Name string
    Role string // "user" or "admin"
}

// Hardcoded users
var users = []User{
    {ID: "luke", Name: "Luke", Role: "user"},
    {ID: "jinny", Name: "Jinny", Role: "user"},
    {ID: "admin", Name: "ADMIN", Role: "admin"},
}
```

#### Products (stored in frontend/backend code)
```go
type Product struct {
    ID          string
    Name        string
    Description string
    Price       int64 // price in cents
}

// Hardcoded products
var products = []Product{
    {ID: "lumaweave", Name: "LumaWeave Reactive Threads", Description: "Smart textile technology", Price: 4999}, // $49.99
    {ID: "coffee-pods", Name: "Atmospheric Coffee Pods", Description: "Premium coffee experience", Price: 2999}, // $29.99
    {ID: "echospout", Name: "EchoSprout Memory Plants", Description: "Living memory storage", Price: 8999}, // $89.99
    {ID: "pocketforge", Name: "PocketForge Nano Printer", Description: "Miniature 3D printing", Price: 19999}, // $199.99
}
```

### Database Schema

#### Transactions Table
```sql
CREATE TABLE transactions (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    product_name TEXT NOT NULL,
    amount INTEGER NOT NULL, -- price in cents
    stripe_session_id TEXT,
    stripe_payment_intent_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, completed, failed, cancelled
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
```

### API Endpoints

#### GET /api/products
Returns the hardcoded list of products
```go
// Response
type ProductsResponse struct {
    Products []Product `json:"products"`
}
```

#### GET /api/users  
Returns the hardcoded list of users
```go
// Response
type UsersResponse struct {
    Users []User `json:"users"`
}
```

#### POST /api/checkout-session
Creates a Stripe checkout session for a product purchase
```go
// Request
type CheckoutSessionRequest struct {
    UserID    string `json:"user_id" binding:"required"`
    ProductID string `json:"product_id" binding:"required"`
}

// Response  
type CheckoutSessionResponse struct {
    SessionID string `json:"session_id"`
    URL       string `json:"url"`
}
```

#### GET /api/transactions/:user_id
Returns transactions for a specific user (for user view)
```go
// Response
type TransactionsResponse struct {
    Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
    ID          string    `json:"id"`
    UserID      string    `json:"user_id"`
    ProductID   string    `json:"product_id"`
    ProductName string    `json:"product_name"`
    Amount      int64     `json:"amount"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### GET /api/transactions
Returns all transactions (for admin view)
```go
// Response - same as above but includes all users' transactions
type AllTransactionsResponse struct {
    Transactions []Transaction `json:"transactions"`
}
```

#### POST /api/webhook
Stripe webhook handler to update transaction status
```go
// Handles Stripe webhook events:
// - checkout.session.completed
// - payment_intent.succeeded
// - payment_intent.payment_failed
```

### Frontend Routes

- `/` - User selector page
- `/user/:user_id` - User dashboard with tabs (products, transactions)  
- `/admin` - Admin dashboard with all transactions

### Session Management

Use browser sessionStorage to maintain current user:
```javascript
// Store selected user
sessionStorage.setItem('currentUser', JSON.stringify(user));

// Retrieve current user
const currentUser = JSON.parse(sessionStorage.getItem('currentUser'));
```

### Stripe Integration

- Use Stripe Checkout for hosted payment flow
- Configure success/cancel URLs to redirect back to application
- Handle webhook events to update transaction status in database
- Store Stripe session ID and payment intent ID for reference 
