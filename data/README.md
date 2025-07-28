# Data Service (Go Personal Data API)

This service is a Go-based HTTP API responsible for managing user personal data, authentication-related information (OTP requests, sessions), and billing details. It interacts with a PostgreSQL database.

## Technologies Used

*   **Go:** The primary language for the service.
*   **PostgreSQL:** The relational database used for data persistence.
*   **`gorilla/mux`:** A powerful HTTP router and URL matcher for building web services.
*   **`lib/pq`:** The PostgreSQL driver for Go's `database/sql` package.
*   **`goose`:** (Implied by `sql/schema` structure) For database migrations.
*   **`sqlc`:** (Implied by `sql/queries` and `internal/db` structure) For generating type-safe Go code from SQL queries.

## Setup

### Environment Variables

Create a `.env` file in the `data/` directory with the following variables:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=energy_sc_bot_db
# Or, alternatively, a full connection string:
# DB_URI="postgres://user:password@localhost:5432/energy_sc_bot_db?sslmode=disable"
```

**Note:** When running with Docker Compose, these values will be provided by the `docker-compose.yml` configuration.

### Database Setup

Ensure a PostgreSQL instance is running and accessible. The database schema is defined in `sql/schema/0001_init.sql`. You can apply migrations using `goose` (if installed) or by manually running the SQL script.

### Build and Run Locally

1.  **Navigate to the service directory:**
    ```bash
    cd data
    ```
2.  **Download Go modules:**
    ```bash
    go mod tidy
    ```
3.  **Run the application:**
    ```bash
    go run main.go
    ```
    The service will start on `http://localhost:8080`.

## API Endpoints (Inferred)

The service exposes HTTP endpoints for managing users and bills. Specific endpoints are defined within `internal/server/server.go` and `internal/db/` queries. Common expected endpoints include:

*   `GET /user/phone/{phone}`: Retrieve user details by phone number.
*   `GET /users/{userID}/bills`: Retrieve bills for a specific user.
*   (Likely) `POST /users`: Create a new user.
*   (Likely) `POST /bills`: Create a new bill.

## Testing

To run unit tests for the data service:

```bash
cd data
go test ./...
```
