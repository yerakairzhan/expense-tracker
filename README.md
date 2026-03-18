# Financial Intelligence Platform

A minimal backend starter project for a financial intelligence system.

## Tech Stack

- **Language**: Go 1.21
- **Framework**: Gin (HTTP framework)
- **Database**: PostgreSQL 15
- **Setup**: sqlc (SQL code generation)
- **Containerization**: Docker & Docker Compose

## Project Structure

```
.
├── cmd/api/                  # Application entry point
├── config/                   # Configuration (database, etc.)
├── db/
│   ├── migrations/          # SQL migration files
│   ├── schema/              # SQL schema (used by sqlc)
│   ├── queries/             # sqlc-generated Go code (package: queries)
│   └── sqlc/                # sqlc query definitions (*.sql)
├── pkg/
│   ├── handler/             # HTTP request handlers
│   ├── models/              # Domain models & DTOs
│   └── repository/          # Data access layer
├── docker-compose.yml       # Docker Compose configuration
├── Dockerfile               # Container build file
├── go.mod                   # Go module definition
└── sqlc.yaml               # sqlc configuration
```

## Features

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/register` | Create a new user |
| POST | `/accounts` | Create a financial account |
| GET | `/transactions` | List transactions for an account |
| GET | `/health` | Health check |

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- PostgreSQL 15+ (if running without Docker)

### Using Docker Compose (Recommended)

```bash
# Build and start containers
docker-compose up --build

# API will be available at http://localhost:8080
```

Swagger UI will be available at `http://localhost:8080/docs`.

### Local Development

```bash
# Install dependencies
go mod download
go mod tidy

# Set environment variables
export DATABASE_URL=postgres://postgres:postgres@localhost:5435/finance_tracker?sslmode=disable
export PORT=8080

# Run migrations (first time only)
psql -h localhost -p 5435 -U postgres -d finance_tracker -f db/migrations/001_init.sql

# Start the server
go run cmd/api/main.go
```

## API Examples

### Register a User

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe"
  }'
```

**Response:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Create an Account

```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "account_type": "savings",
    "balance": "10000.00",
    "currency": "USD"
  }'
```

**Response:**
```json
{
  "id": 1,
  "user_id": 1,
  "account_type": "savings",
  "balance": "10000.00",
  "currency": "USD",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### List Transactions

```bash
curl "http://localhost:8080/transactions?account_id=1&limit=50&offset=0"
```

**Response:**
```json
[
  {
    "id": 1,
    "account_id": 1,
    "amount": "100.50",
    "description": "Deposit",
    "transaction_type": "credit",
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

## Database Schema

### users
- `id` (INT, PK)
- `email` (VARCHAR, UNIQUE)
- `name` (VARCHAR)
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

### accounts
- `id` (INT, PK)
- `user_id` (INT, FK → users)
- `account_type` (VARCHAR)
- `balance` (DECIMAL)
- `currency` (VARCHAR)
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

### transactions
- `id` (INT, PK)
- `account_id` (INT, FK → accounts)
- `amount` (DECIMAL)
- `description` (VARCHAR)
- `transaction_type` (VARCHAR)
- `created_at` (TIMESTAMP)

## Setting Up sqlc

sqlc generates type-safe Go code from SQL queries. To use it:

```bash
# Install sqlc (if not already installed)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate code from SQL queries
sqlc generate

# Generated code will be in db/queries/
```

Note: The repository layer uses the generated `db/queries` package (created by sqlc from `db/sqlc/*.sql`).

## Development Workflow

### 1. Add New Endpoint

1. Create handler in `pkg/handler/`
2. Create repository method in `pkg/repository/`
3. Add route in `cmd/api/main.go`
4. Test with curl or Postman

### 2. Modify Database Schema

1. Create new migration in `db/migrations/`
2. Update sqlc queries in `db/sqlc/*.sql`
3. Run `sqlc generate`
4. Update models in `pkg/models/`

### 3. Add Business Logic

1. Create service layer in `pkg/service/` (future)
2. Call from handlers
3. Keep repositories focused on data access only

## Future Enhancements

- [ ] Authentication & Authorization
- [ ] Service layer for business logic
- [ ] Middleware for logging, error handling
- [ ] Transaction management
- [ ] Validation utilities
- [ ] Error handling improvements
- [ ] Integration tests
- [ ] API documentation (Swagger/OpenAPI)
- [ ] Caching layer (Redis)
- [ ] Background job workers
- [ ] Monitoring & observability

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker-compose ps

# View logs
docker-compose logs postgres

# Restart services
docker-compose down && docker-compose up --build
```

### Port Already in Use

```bash
# Change ports in docker-compose.yml or use different port
docker-compose up -p 8081:8080
```

## License

MIT
