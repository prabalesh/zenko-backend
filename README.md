# Zenko Backend

Multiplayer brain duels mobile app backend.

## Tech Stack
- **Go 1.22+**
- **Fiber v2** (Web Framework)
- **PostgreSQL 16** (Primary DB)
- **Redis 7** (Caching/Socket state)
- **sqlc** (Type-safe SQL)
- **golang-migrate** (Database migrations)
- **Zerolog** (Structured logging)
- **Viper** (Configuration)

## Getting Started

### Prerequisites
- Docker & Docker Compose
- Go 1.22+
- Make

### Setup
1. Clone the repository.
2. Create your `.env` file:
   ```bash
   cp .env.example .env
   ```
3. Update the `.env` file with your secrets (especially `JWT_SECRET`).

### Running Locally
To start the entire stack (API, Postgres, Redis):
```bash
make dev
```

To run the Go application directly:
```bash
make run
```

### Database Migrations
To run migrations up:
```bash
make migrate-up
```

### Development Commands
- `make build`: Build the server binary.
- `make test`: Run tests.
- `make migrate-down`: Rollback migrations.
