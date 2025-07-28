# Energy SC Bot

## Architecture

This project consists of three microservices designed to work together:

1.  **Go Telegram Bot (`tg-bot`)**
    *   Responds to Telegram users.
    *   Stores the last 10 messages of each user in Redis for context.
    *   Manages dialogue logic through states.
    *   Integrates with the RAG service via gRPC.
    *   Retrieves personal data via an HTTP API.

2.  **Python RAG Service (`rag-service`)**
    *   Processes text messages: classifies, enhances, and answers using an LLM (Large Language Model) and a knowledge base.
    *   Acts as a gRPC server, implementing methods defined in `proto/rag.proto`.

3.  **Go Personal Data API (`data`)**
    *   An HTTP service that provides user personal data, bills, history, and other information from a PostgreSQL database.

### Technologies Used

*   **Go:** Primary language for the Telegram bot and Personal Data API.
*   **Python:** Used for the RAG service (LLM + knowledge base).
*   **Redis:** For storing message history and dialogue state.
*   **PostgreSQL:** For persistent storage of user data.
*   **gRPC:** For high-performance communication between the bot and the RAG service.
*   **Docker/Docker Compose:** For building, orchestrating, and running the entire system.

## Quick Start (Docker Compose)

To get all services up and running quickly using Docker Compose:

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/goinginblind/energy-sc-bot.git
    cd energy-sc-bot
    ```

2.  **Create `.env` files:**
    Create `.env` files in the root directory and within each service directory (`tg-bot/`, `rag-service/`, `data/`) based on the examples provided below and in their respective `README.md` files.

    **Example `.env` for `tg-bot` (in `tg-bot/.env`):**
    ```env
    TELEGRAM_TOKEN=your-telegram-bot-token
    REDIS_ADDR=redis:6379
    REDIS_PASSWORD=
    REDIS_DB=0
    GRPC_SERVICE_ADDR=rag-service:50051
    PERSONAL_DATA_API_URL=http://data:8080
    METRICS_PORT=8081
    ```

    **Example `.env` for `rag-service` (in `rag-service/.env`):**
    ```env
    # No specific environment variables needed for basic operation, but can be added for LLM API keys etc.
    ```

    **Example `.env` for `data` (in `data/.env`):**
    ```env
    DB_HOST=db
    DB_PORT=5432
    DB_USER=user
    DB_PASSWORD=password
    DB_NAME=energy_sc_bot_db
    ```

3.  **Start the services:**
    From the root of the project, run:
    ```bash
    docker-compose up --build
    ```
    This will build the Docker images for all services and start them, along with Redis and PostgreSQL.

## Development Setup (Running Services Individually)

If you prefer to run services locally for development:

1.  **Start dependencies:**
    Ensure Redis and PostgreSQL are running (e.g., via `docker-compose up redis db` or locally installed instances).

2.  **Set up `tg-bot` (Go):**
    ```bash
    cd tg-bot
    go mod tidy
    # Create .env file as per example above
    go run cmd/app/main.go
    ```

3.  **Set up `rag-service` (Python):**
    ```bash
    cd rag-service
    pip install -r requirements.txt
    # Create .env file if needed
    python main.py
    ```

4.  **Set up `data` (Go):**
    ```bash
    cd data
    go mod tidy
    # Create .env file as per example above
    go run main.go
    ```

## Testing

To run tests for all services:

*   **For Go services (`tg-bot`, `data`):**
    ```bash
    cd tg-bot && go test ./...
    cd ../data && go test ./...
    ```

*   **For Python service (`rag-service`):**
    ```bash
    cd rag-service && pytest
    ```

## Services

*   [`tg-bot/`](tg-bot/README.md) — Go Telegram Bot source
*   [`rag-service/`](rag-service/README.md) — Python RAG Service source
*   [`data/`](data/README.md) — Go Personal Data API source

## License

MIT