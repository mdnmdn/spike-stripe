# Project context

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a minimal Go HTTP API skeleton designed for spiking Stripe payment integrations. It uses the Gin web framework and provides mock payment endpoints that can be replaced with actual Stripe integration during development.

## Architecture

- **Main application**: `cmd/server/main.go` - Entry point that initializes the payments service and starts the Gin server
- **API layer**: `internal/api/` - Contains HTTP handlers and routing logic using Gin framework
    - `router.go` - Sets up routes and serves embedded frontend assets
    - `handlers.go` - HTTP endpoint implementations
- **Business logic**: `internal/payments/` - Payment service abstraction
    - `service.go` - Mock payment service that mimics Stripe operations
- **Frontend**: `cmd/server/frontend/` - Static HTML assets embedded into binary
- **API specification**: `openapi.yaml` - OpenAPI 3.0 spec defining the REST endpoints
- **Database**: `internal/db/` helpers; SQLc generated code goes here. Migrations in `db/migrations`, queries in `db/queries`. Migration runner at `cmd/migrate/`. Uses SQLite locally and Turso when `TURSO_DATABASE_URL` is set.

The service layer pattern allows easy swapping between mock and real Stripe implementations during development.

## Development Commands

### Build and Run
```bash
# Build the application
go build -o stripe-go-spike cmd/server/main.go

# Run with environment variables
STRIPE_SECRET_KEY=sk_test_xxx STRIPE_PUBLISHABLE_KEY=pk_test_xxx ./stripe-go-spike

# Run without Stripe keys (uses mock service)
./stripe-go-spike
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run specific package tests
go test ./internal/api
```

### Docker
```bash
# Build container
docker build -t stripe-go-spike .

# Run container
docker run -p 8060:8060 stripe-go-spike
```

## Environment Variables

The server supports dotenv files. On startup, it loads the following files in order (later files are lower precedence; OS environment always wins):

1. `.env.{APP_ENV}.local` (or `.env.{GO_ENV}.local`)
2. `.env.local`
3. `.env.{APP_ENV}` (or `.env.{GO_ENV}`)
4. `.env`

Notes:
- Variables already set in the OS environment are never overwritten by file values.
- If `APP_ENV` is unset, `GO_ENV` is used. If both are unset, only `.env.local` and `.env` are considered.

Available variables:
- `DB_PATH` - Optional path to local SQLite file (default: `_data/db-spike-strip.sqlite3`)
- `TURSO_DATABASE_URL` - Optional: when set, connect to Turso/LibSQL
- `TURSO_AUTH_TOKEN` - Optional token for Turso
- `STRIPE_SECRET_KEY` - Stripe secret key (optional for mock)
- `STRIPE_PUBLISHABLE_KEY` - Stripe publishable key (optional for mock)
- `STRIPE_WEBHOOK_SECRET` - Webhook endpoint secret (optional for mock)
- `PORT` - Override server port (default: 8060)
- `APP_ADDR` - Override server address (default: :8060)

## API Endpoints

- `GET /api/health` - Health check endpoint
- `POST /api/checkout-session` - Create mock/real Stripe checkout session
- `POST /api/webhook` - Receive Stripe webhook events
- `GET /` - Redirects to frontend at `/app`
- `GET /app` - Serves embedded frontend assets

The application serves both API endpoints and a basic frontend from the same server.