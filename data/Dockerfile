FROM golang:1.21-alpine AS builder
WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
RUN go build -o data-service ./main.go

# Финальный образ
FROM alpine:latest
WORKDIR /app

# Устанавливаем зависимости для работы с PSQL
RUN apk add --no-cache ca-certificates

# Копируем бинарник
COPY --from=builder /app/data-service ./data-service

EXPOSE 8080

CMD ["./data-service"]
