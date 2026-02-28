## 1. Module and Dependencies

- [x] 1.1 Update `backend/go.mod` — keep module as `backend`, add all core dependencies (chi/v5, pgx/v5, shopspring/decimal, golang-jwt/v5, caarlos0/env/v11, golang.org/x/oauth2), run `go mod tidy`
- [x] 1.2 Delete `backend/main.go` (GoLand scaffold)

## 2. Directory Structure

- [x] 2.1 Create all directories: `cmd/api/`, `model/`, `store/`, `handler/`, `service/`, `auth/`, `config/`, `migrations/`, `db/queries/`

## 3. Domain Value Objects

- [x] 3.1 Create `model/money.go` — Money type with unexported fields (amount decimal.Decimal, currency string), NewMoney, NewMoneyFromString, Zero constructors, Add/Sub/Equal/GreaterThan/LessThanOrEqual/IsZero/IsNegative/IsPositive methods (value receivers), String() with 2 decimal places, mustMatchCurrency panic helper
- [x] 3.2 Create `model/money_test.go` — table-driven tests for all Money operations: creation with default/explicit currency, string parsing (valid/invalid), arithmetic (add/sub/zero), currency mismatch panic, comparisons, string formatting
- [x] 3.3 Create `model/daterange.go` — DateRange type with unexported fields (from, to time.Time), NewDateRange (validates from < to), MonthRange, WeekRange (ISO week), DayRange convenience constructors, From()/To() accessors
- [x] 3.4 Create `model/daterange_test.go` — table-driven tests for all DateRange operations: valid/invalid ranges, same-date rejection, MonthRange/WeekRange/DayRange constructors, accessor methods
- [x] 3.5 Create `model/errors.go` — sentinel errors for all domain areas (auth, workspace, transaction, category) using errors.New

## 4. Configuration

- [x] 4.1 Create `config/config.go` — Config, ServerConfig, DatabaseConfig, AuthConfig structs with caarlos0/env struct tags and defaults, DSN() method on DatabaseConfig, Load() function
- [x] 4.2 Create `.env.example` — all configuration variables with safe defaults (server, database, JWT, Google OAuth, GitHub OAuth)

## 5. PostgreSQL Connection

- [x] 5.1 Create `store/db.go` — NewPool(ctx, DatabaseConfig) function that parses DSN, creates pgxpool with MaxConns, pings the database

## 6. HTTP Server Skeleton

- [x] 6.1 Create `handler/response.go` — envelope type, respondJSON(w, status, data) and respondError(w, status, message) helpers
- [x] 6.2 Create `handler/request.go` — decodeJSON(r, dst) with 1MB MaxBytesReader and DisallowUnknownFields
- [x] 6.3 Create `handler/router.go` — NewRouter() function returning chi.Router with global middleware (RequestID, RealIP, Logger, Recoverer), GET /healthz endpoint
- [x] 6.4 Create `cmd/api/main.go` — main() calls run(), run() loads config, creates DB pool, builds router, starts http.Server with configured timeouts, graceful shutdown on SIGINT/SIGTERM with 10s deadline, slog structured logging

## 7. Build Tooling

- [x] 7.1 Create `Makefile` — targets: build, run, test, lint, sqlc, migrate-up, migrate-down, migrate-create, tidy
- [x] 7.2 Create `.golangci.yml` — enable linters: govet, errcheck, staticcheck, gosimple, ineffassign, goconst, gocyclo, unused, gosec
- [x] 7.3 Create `sqlc.yaml` — PostgreSQL engine, pgx/v5 sql_package, NUMERIC→decimal override, emit_json_tags, emit_empty_slices, emit_interface, schema from migrations/, queries from db/queries/

## 8. Documentation

- [x] 8.1 Update `CLAUDE.md` — reflect new directory structure, updated build/run/test commands, architectural overview

## 9. Verification

- [x] 9.1 Run `go build ./...` — must compile with zero errors
- [x] 9.2 Run `go vet ./...` — must pass with zero warnings
- [x] 9.3 Run `go test -race -count=1 ./...` — all tests must pass
