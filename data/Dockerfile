# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install git for go mod (if needed)
RUN apk add --no-cache git

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code and build
COPY . .
RUN go build -o data-service ./main.go

# Final stage
FROM alpine:latest
WORKDIR /app

# Install CA certs for HTTPS (PostgreSQL/TLS)
RUN apk add --no-cache ca-certificates

# Copy the binary from builder
COPY --from=builder /app/data-service .

EXPOSE 8080
CMD ["./data-service"]
