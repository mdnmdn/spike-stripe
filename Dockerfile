# Simple multi-stage Dockerfile for a Go web app
# Build stage
FROM golang:1.24-alpine AS build
WORKDIR /src

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build a static binary
ENV CGO_ENABLED=0
RUN go build -o /out/stripe-go-spike ./cmd/server

# Runtime stage
FROM alpine:3.19

# Create non-root user and prepare app directory (writable for DB _data)
RUN addgroup -S appgroup \
  && adduser -S appuser -G appgroup \
  && mkdir -p /app/_data \
  && chown -R appuser:appgroup /app

WORKDIR /app

# Copy binary
COPY --from=build /out/stripe-go-spike ./stripe-go-spike

# Environment
ENV RUN_MIGRATION=true

# Expose HTTP port
EXPOSE 8060

# Run as non-root
USER appuser

# Start the server
CMD ["./stripe-go-spike"]
