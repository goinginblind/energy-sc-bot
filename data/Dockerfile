# 1. Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code for the data service
COPY . .

# Build the application
# -o /app/main creates the binary in the /app directory
# CGO_ENABLED=0 is important for creating a static binary
# -ldflags="-s -w" strips debug information to reduce binary size
RUN CGO_ENABLED=0 go build -o /app/main -ldflags="-s -w" ./main.go

# 2. Final stage
FROM alpine:latest

# Add non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy SQL schema and queries
COPY sql ./sql
COPY internal ./internal

# Set user
USER appuser

# Expose port and run
EXPOSE 8080
CMD ["./main"]
