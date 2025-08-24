# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
ENV CGO_ENABLED=0
RUN go build -o /stripe-go-spike cmd/server/main.go

# Stage 2: Create the final image
FROM alpine:3.19

# Copy the built binary from the builder stage
COPY --from=builder /stripe-go-spike /stripe-go-spike

# Create a non-root user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Expose the port the app runs on
EXPOSE 8060

# Set the command to run the application
CMD ["/stripe-go-spike"]
