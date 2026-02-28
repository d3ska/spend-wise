## Context

The `backend/` directory contains a GoLand scaffold `main.go` (tutorial code) and a bare `go.mod` declaring `module backend` with Go 1.26. No directory structure, no dependencies, no build tooling. This change establishes the entire project foundation that all subsequent changes build upon.

The developer is an experienced Java engineer transitioning to Go. The design prioritizes idiomatic Go patterns over Java-style abstractions (no DI containers, no getter/setter boilerplate, no generic Repository<T>, no exception hierarchies).

## Goals / Non-Goals

**Goals:**
- Establish pragmatic flat package layout (cmd/api, model, store, handler, service, auth, config)
- Create domain value objects (Money, DateRange) with full test coverage
- Set up a running HTTP server with health check endpoint
- Configure build tooling (Makefile, golangci-lint, sqlc)
- Set up PostgreSQL connection pool and configuration management
- Define sentinel domain errors used across all future changes

**Non-Goals:**
- No business entities (User, Workspace, Transaction) — those belong to subsequent changes
- No database migrations or sqlc queries — only the sqlc config and connection pool setup
- No authentication or authorization — added in `s002-auth-and-users`
- No repository interfaces yet — each change defines its own as needed
- No external API integrations

## Decisions

### 1. Module path: `backend`

Keep the simple `module backend` path. This is a standalone application, not a library that will be imported by other Go modules. A GitHub-based module path can be adopted later if needed (e.g., for open-sourcing).

### 2. Pragmatic flat package layout (no `internal/`)

Directory structure uses flat packages directly under `backend/`: `model/`, `store/`, `handler/`, `service/`, `auth/`, `config/`. No `internal/` wrapper — unnecessary for a single-module web app that won't be imported as a library. One-way dependency chain: `model ← store ← service ← handler ← cmd/api`, with `auth` feeding into `handler`.

**Alternative considered:** Clean Architecture with `internal/domain/`, `internal/app/`, `internal/adapter/`. Rejected as over-engineered for this project — adds 2 levels of nesting for no benefit. The flat layout is inspired by Anthony GG's pragmatic Go project structure, keeping packages discoverable and dependency direction clear without layered abstractions.

### 3. `shopspring/decimal` for Money, not `int64` cents

Money value object wraps `decimal.Decimal` for arbitrary-precision arithmetic. Using int64 cents would require manual scaling and lose sub-cent precision needed for splits.

**Alternative considered:** `int64` representing cents. Simpler but loses precision on division/splits and requires manual formatting everywhere.

### 4. `caarlos0/env` for configuration, not Viper

Struct-tag-based env loading is simpler and more idiomatic for 12-factor apps. Viper adds YAML/TOML/remote config complexity that isn't needed.

**Alternative considered:** `spf13/viper`. Feature-rich but heavy, brings reflection-based access patterns similar to Spring's `@ConfigurationProperties`. Overkill for env-var-only config.

### 5. `log/slog` (stdlib), not zap or logrus

Go 1.21+ includes `log/slog` for structured logging. No external dependency needed. If performance becomes critical, swap the handler — the API stays the same.

### 6. Panic on Money currency mismatch

`Money.Add(other)` panics if currencies differ. This is a programmer error (calling code should validate), not a runtime condition. Services validate currencies before arithmetic.

**Alternative considered:** Return `(Money, error)` from all arithmetic. Makes every expression verbose (`sum, err := a.Add(b); if err != nil...`) for a condition that indicates a bug, not user input.

## Risks / Trade-offs

- **[Risk] Module rename breaks IDE references** → The only existing code is the scaffold `main.go` which is deleted. No actual references to break.
- **[Risk] Dependencies pinned at specific versions may drift** → `go.mod` uses minimum versions. Run `go get -u` periodically.
- **[Risk] golangci-lint not installed on developer machine** → Makefile `lint` target will fail. Document installation in CLAUDE.md.
- **[Trade-off] Money panics vs errors** → Panics give cleaner arithmetic API but require discipline at the service layer to validate currencies before domain operations.
