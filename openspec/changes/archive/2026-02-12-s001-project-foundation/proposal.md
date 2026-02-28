## Why

The `backend/` directory contains only a GoLand scaffold `main.go` and a bare `go.mod`. Before any feature work can begin, the project needs a proper Go project structure, build tooling, domain value objects, and a running HTTP server skeleton. Without this foundation, every subsequent change would need to simultaneously establish conventions, leading to inconsistency.

## What Changes

- Replace the GoLand scaffold `main.go` with a proper `cmd/api/main.go` entry point using a pragmatic flat package layout
- Establish the full directory structure: `cmd/api/`, `model/`, `store/`, `handler/`, `service/`, `auth/`, `config/`
- Update `go.mod` with core dependencies (Chi router, pgx, shopspring/decimal, jwt, caarlos0/env, oauth2)
- Create the `Money` immutable value object (wraps `shopspring/decimal`) with currency support and arithmetic operations
- Create the `DateRange` value object with `from < to` enforcement and convenience constructors (month, week, day)
- Create `model/errors.go` with sentinel errors for all domain areas
- Create `config/config.go` with env-based configuration loading (server, database, JWT, OAuth)
- Create a minimal Chi HTTP server with `GET /healthz` endpoint, structured logging (`log/slog`), and graceful shutdown
- Create HTTP response/request helpers (`handler/response.go`, `handler/request.go`) with JSON envelope pattern and 1MB body limit
- Create `Makefile` with build, run, test, lint, migration, and sqlc targets
- Create `.golangci.yml` with linters: govet, errcheck, staticcheck, gosimple, ineffassign, goconst, gocyclo, unused, gosec
- Create `.env.example` with all configuration variables
- Create `sqlc.yaml` with PostgreSQL + pgx/v5 config, NUMERIC→decimal override, and empty slice emission
- Create `store/db.go` with pgx connection pool constructor
- Delete the GoLand scaffold `backend/main.go`
- Update `CLAUDE.md` with the new project structure and commands

## Capabilities

### New Capabilities
- `project-structure`: Go project layout with flat packages (cmd/api, model, store, handler, service, auth, config), build tooling (Makefile, golangci-lint, sqlc config), and configuration management
- `domain-value-objects`: Money value object (immutable, arbitrary-precision via shopspring/decimal, currency-aware arithmetic) and DateRange value object (from/to enforcement, month/week/day constructors)
- `http-server`: Chi-based HTTP server skeleton with healthz endpoint, graceful shutdown, structured logging, JSON response helpers, and request decoding with size limits

### Modified Capabilities

## Impact

- **Code:** Replaces `backend/main.go`, creates ~15 new files establishing the full project skeleton
- **Dependencies:** Adds chi/v5, pgx/v5, shopspring/decimal, golang-jwt/v5, caarlos0/env/v11, golang.org/x/oauth2
- **Build:** New Makefile with `make build`, `make run`, `make test`, `make lint`, `make sqlc`, `make migrate-*`
- **Verification:** `go build ./...` compiles, `go vet ./...` passes, `GET /healthz` returns `{"status":"ok"}`
