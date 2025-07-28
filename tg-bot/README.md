# TG Bot Service (Go Telegram Bot)

This service is the main Telegram bot interface for the Energy SC Bot project. It handles user interactions, manages dialogue states, stores message history in Redis, and communicates with the RAG (Retrieval-Augmented Generation) service via gRPC and the Personal Data API via HTTP. It also includes structured logging with Zap and exposes Prometheus metrics for observability.

## Technologies Used

*   **Go:** The primary language for the bot.
*   **`go-telegram-bot-api/telegram-bot-api/v5`:** For interacting with the Telegram Bot API.
*   **Redis:** Used for storing user session states and message history.
*   **gRPC:** For communication with the RAG service.
*   **HTTP Client:** For communication with the Personal Data API.
*   **`go.uber.org/zap`:** For structured logging.
*   **`prometheus/client_golang`:** For exposing Prometheus metrics.

## Setup

### Environment Variables

Create a `.env` file in the `tg-bot/` directory with the following variables:

```env
TELEGRAM_TOKEN=your-telegram-bot-token
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
GRPC_SERVICE_ADDR=rag-service:50051
PERSONAL_DATA_API_URL=http://data:8080
METRICS_PORT=8081
```

**Note:** When running with Docker Compose, these values will be provided by the `docker-compose.yml` configuration.

### Dependencies

Ensure Redis, the RAG service, and the Data service are running and accessible.

## Build and Run Locally

1.  **Navigate to the service directory:**
    ```bash
    cd tg-bot
    ```
2.  **Download Go modules:**
    ```bash
    go mod tidy
    ```
3.  **Run the application:**
    ```bash
    go run cmd/app/main.go
    ```
    The bot will start polling for updates from Telegram. A metrics endpoint will be available at `http://localhost:8081/metrics`.

## Key Functionalities

*   **User Authentication:** Handles login flow with OTP (One-Time Password).
*   **Dialogue Management:** Manages user states (e.g., `StateStart`, `StateAwaitingLoginInput`, `StateGeneralInquiry`).
*   **Message History:** Stores recent user messages in Redis for context.
*   **General Inquiry:** Routes user questions to the RAG service for AI-powered answers.
*   **Personal Data Access:** Fetches user-specific data (like bills) from the Data service to enrich RAG queries.
*   **Observability:** Integrates Zap for structured logging and Prometheus for metrics collection.

## Testing

To run unit tests for the Telegram bot service:

```bash
cd tg-bot
go test ./...
```