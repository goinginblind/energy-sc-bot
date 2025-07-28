# 1. Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -o /app/main -ldflags="-s -w" ./cmd/app/main.go

# 2. Final stage
FROM alpine:latest

# Add non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Set user
USER appuser

# Run the bot
CMD ["./main"]