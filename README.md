# Telegram Bot + RAG + Personal Data

## Архитектура

Проект состоит из трёх микросервисов:

1. **Go Telegram Bot**  
   - Отвечает пользователям Telegram
   - Хранит последние 10 сообщений каждого пользователя в Redis
   - Управляет логикой диалога через стейты
   - Интегрируется с RAG-сервисом через gRPC
   - Получает персональные данные через HTTP API

2. **Python RAG**  
   - Обрабатывает текстовые сообщения: классифицирует, улучшает и отвечает с помощью LLM + базы знаний
   - gRPC сервер, реализующий методы из `proto/rag.proto`

3. **Go Personal Data API**  
   - HTTP-сервис, возвращающий персональные данные пользователя, счета, историю и т.п. из PostgreSQL

### Используемые технологии

- **Go** — основной язык для Telegram-бота и API персональных данных
- **Python** — для RAG-сервиса (LLM + база знаний)
- **Redis** — хранение истории сообщений и состояния диалога
- **PostgreSQL** — хранение пользовательских данных
- **gRPC** — взаимодействие между ботом и RAG-сервисом
- **Docker/Docker Compose** — сборка и запуск всей системы

## Быстрый старт

1. Клонируйте репозиторий и создайте `.env` файлы для сервисов.
2. Запустите через Docker Compose (пример будет добавлен позже).

## Сервисы

- `tg-bot/` — исходники Telegram-бота (Go)
- `rag-service/` — исходники RAG-сервиса (Python)
- `data/` — исходники API персональных данных (Go)

## Пример переменных окружения для Telegram Bot

```
TELEGRAM_TOKEN=your-telegram-bot-token
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
GRPC_SERVICE_ADDR=rag:50051
PERSONAL_DATA_API_URL=http://personal-data-api:8080
```

## Запуск Telegram Bot вручную

```sh
cd tg-bot/cmd/app
go build -o bot
./bot
```

## Запуск через Docker

```sh
docker build -t energy-sc-bot tg-bot/
docker run --env-file tg-bot/.env energy-sc-bot
```

## Лицензия

MIT
