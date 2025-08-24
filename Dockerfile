# Stage 1: Build the Go application
FROM golang:1.24.3-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN go build -o /pa11y-go-server cmd/server/main.go

# Stage 2: Create the final image
FROM node:20-slim

# Install pa11y globally
RUN npm install -g pa11y

# Copy the built binary from the builder stage
COPY --from=builder /pa11y-go-server /pa11y-go-server

# Expose the port the app runs on
EXPOSE 8080

# Set the command to run the application
CMD ["/pa11y-go-server"]
