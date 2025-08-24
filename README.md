# stripe-go-spike

A tiny Go API skeaeton to spike a Stripe integration. It exposes minimal endpoints to create a mock Checkout Session and receive webhooks. Replace the mock payments service with stripe-go as you iterate.

## Installation

This project now includes optional SQLc + SQLite/Turso setup for a small cache table. See _docs/sqlc.md for details.


### Prerequisites

Before you can use this tool, you need to have the following installed on your system:

*   **Go**: This tool is written in Go, so you'll need to have the Go toolchain installed. You can find installation instructions on the [official Go website](https://golang.org/doc/install).
- Stripe account (optional for now). When ready, set env vars:
  - STRIPE_SECRET_KEY
  - STRIPE_PUBLISHABLE_KEY
  - STRIPE_WEBHOOK_SECRET

### Building the server

```bash
go build -o stripe-go-spike cmd/server/main.go
```

## Usage

See _docs/sqlc.md for database and migrations. The HTTP API remains unchanged.

Dotenv support: on startup the server loads environment from these files in order (earlier = higher precedence; OS env always wins):
1. `.env.{APP_ENV}.local` (or `.env.{GO_ENV}.local`)
2. `.env.local`
3. `.env.{APP_ENV}` (or `.env.{GO_ENV}`)
4. `.env`

Examples:
```bash
# using dotenv files
APP_ENV=development ./stripe-go-spike

# passing env directly
STRIPE_SECRET_KEY=sk_test_xxx STRIPE_PUBLISHABLE_KEY=pk_test_xxx STRIPE_WEBHOOK_SECRET=whsec_xxx ./stripe-go-spike
```

The server starts on port 8060 by default; override with PORT or APP_ADDR. Database settings are optional and only used if you adopt the SQLc integration.

Run migrations automatically on startup by setting RUN_MIGRATION=true (or any truthy value like 1, yes, t). This connects using TURSO_DATABASE_URL/TURSO_AUTH_TOKEN if set, otherwise a local SQLite file (DB_PATH or default path).

## API

The server exposes minimal endpoints:

- GET /api/health
- POST /api/checkout-session
- POST /api/webhook




