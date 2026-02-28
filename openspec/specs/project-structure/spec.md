## ADDED Requirements

### Requirement: Pragmatic flat package layout
The project SHALL use a flat package layout with clear one-way dependency direction: `model ← store ← service ← handler ← cmd/api`, with `auth` feeding into `handler`. No `internal/` directory.

#### Scenario: Directory structure exists after initialization
- **WHEN** the project foundation is created
- **THEN** the following directories SHALL exist:
  - `backend/cmd/api/` — application entry point and composition root
  - `backend/model/` — entities, value objects, sentinel errors
  - `backend/store/` — pgx connection pool, sqlc generated code, repository implementations
  - `backend/handler/` — HTTP handlers, response/request helpers, handler-local DTOs
  - `backend/service/` — business logic (rules engine, summary, import)
  - `backend/auth/` — OAuth2 provider adapters, JWT middleware
  - `backend/config/` — environment-based configuration loading
  - `backend/migrations/` — database migration files
  - `backend/db/queries/` — sqlc SQL query files

### Requirement: Go module with dependencies
The `go.mod` file SHALL declare the module as `backend` with Go 1.26 and SHALL include all core dependencies.

#### Scenario: Module compiles with dependencies
- **WHEN** `go build ./...` is run from the `backend/` directory
- **THEN** the build SHALL succeed with zero errors
- **AND** the following dependencies SHALL be resolvable: `github.com/go-chi/chi/v5`, `github.com/jackc/pgx/v5`, `github.com/shopspring/decimal`, `github.com/golang-jwt/jwt/v5`, `github.com/caarlos0/env/v11`, `golang.org/x/oauth2`

### Requirement: GoLand scaffold removed
The GoLand-generated `backend/main.go` scaffold SHALL be deleted and replaced by `backend/cmd/api/main.go`.

#### Scenario: Old scaffold does not exist
- **WHEN** the project foundation is created
- **THEN** `backend/main.go` SHALL NOT exist
- **AND** `backend/cmd/api/main.go` SHALL be the sole entry point

### Requirement: Makefile with standard targets
A `Makefile` SHALL provide targets for all common development operations.

#### Scenario: All Makefile targets are functional
- **WHEN** the developer uses the Makefile
- **THEN** the following targets SHALL exist and function:
  - `make build` — compiles to `bin/spend-wise-api`
  - `make run` — runs the API server via `go run ./cmd/api`
  - `make test` — runs `go test -race -count=1 ./...`
  - `make lint` — runs `golangci-lint run`
  - `make sqlc` — runs `sqlc generate`
  - `make migrate-up` — applies all pending migrations
  - `make migrate-down` — rolls back last migration
  - `make tidy` — runs `go mod tidy`

### Requirement: Linter configuration
A `.golangci.yml` file SHALL configure golangci-lint with production-grade linters.

#### Scenario: Linter config enables required analyzers
- **WHEN** `golangci-lint run` is executed
- **THEN** the following linters SHALL be enabled: `govet`, `errcheck`, `staticcheck`, `gosimple`, `ineffassign`, `goconst`, `gocyclo`, `unused`, `gosec`

### Requirement: sqlc configuration
A `sqlc.yaml` file SHALL configure sqlc for PostgreSQL with pgx/v5, mapping `NUMERIC` to `shopspring/decimal.Decimal`.

#### Scenario: sqlc config is valid
- **WHEN** `sqlc generate` is run (with migration files present)
- **THEN** generated code SHALL use `pgx/v5` as the SQL package
- **AND** `NUMERIC` database columns SHALL map to `github.com/shopspring/decimal.Decimal`
- **AND** empty query results SHALL return `[]T{}` (not nil)
- **AND** generated types SHALL include JSON struct tags

### Requirement: Environment configuration example
A `.env.example` file SHALL document all configuration variables with safe defaults.

#### Scenario: All config variables documented
- **WHEN** a developer copies `.env.example` to `.env`
- **THEN** the file SHALL contain variables for: server (host, port, timeouts), database (host, port, user, password, name, sslmode, max_conns), JWT (secret, expiry), OAuth providers (Google and GitHub client IDs, secrets, redirect URLs)

### Requirement: Configuration loading from environment
The `config` package SHALL load all configuration from environment variables with sensible defaults.

#### Scenario: Config loads with defaults
- **WHEN** no environment variables are set
- **THEN** `config.Load()` SHALL return a valid `Config` with default values (server on 0.0.0.0:8080, database on localhost:5432, etc.)

#### Scenario: Config loads from environment
- **WHEN** `SERVER_PORT=9090` is set
- **THEN** `config.Load()` SHALL return a `Config` with `Server.Port == 9090`

#### Scenario: Database DSN generation
- **WHEN** `DatabaseConfig.DSN()` is called
- **THEN** it SHALL return a valid PostgreSQL connection string in the format `postgres://user:pass@host:port/dbname?sslmode=X`

### Requirement: PostgreSQL connection pool
The `store` package SHALL provide a function to create a pgx connection pool from config.

#### Scenario: Pool creation and ping
- **WHEN** `store.NewPool(ctx, cfg)` is called with valid database config
- **THEN** it SHALL return a `*pgxpool.Pool` that has been pinged successfully
- **AND** the pool SHALL respect `MaxConns` from config
