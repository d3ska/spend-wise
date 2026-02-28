# SpendWise

Collaborative household financial ledger. Track shared expenses, split costs, and view spending summaries.

## Quick Start

```bash
docker compose up
```

This starts all services:

| Service    | URL                    | Description              |
|------------|------------------------|--------------------------|
| Frontend   | https://localhost:3000 | React app                |
| API        | http://localhost:8080  | Go backend               |
| PostgreSQL | localhost:5432         | Database                 |
| pgAdmin    | http://localhost:5050  | Database UI              |

## pgAdmin

Open http://localhost:5050 and log in:

- **Email:** `admin@spendwise.dev`
- **Password:** `admin`

To connect to the database, add a new server:

- **Host:** `db`
- **Port:** `5432`
- **Username:** `postgres`
- **Password:** `postgres`
- **Database:** `spendwise`

## Development (without Docker)

All backend commands run from the `backend/` directory:

```bash
# Run the API server
make run

# Build binary
make build

# Run tests
make test

# Vet/lint
go vet ./...
make lint

# Tidy dependencies
make tidy

# Database migrations (requires DATABASE_URL)
make migrate-up
make migrate-down

# Generate sqlc code
make sqlc
```

Frontend:

```bash
cd frontend
npm run dev
```
