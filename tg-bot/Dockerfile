# Stage 1: Build the binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git (required for go mod sometimes)
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bot

# Stage 2: Slim runtime image
FROM alpine:latest

WORKDIR /app

# Copy built binary
COPY --from=builder /app/bot .

# Copy env if needed
COPY .env .env

# Smallest possible base, just enough for networking
RUN apk add --no-cache ca-certificates

CMD ["./bot"]
