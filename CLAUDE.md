# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SpendWise is a collaborative household financial ledger. Users track shared expenses, split costs, and view spending summaries across configurable date ranges. Authentication is via OAuth2 SSO (Google, GitHub).

## Tech Stack

- **Language:** Go 1.26
- **Module:** `backend` (defined in `backend/go.mod`)
- **Router:** go-chi/chi/v5
- **Database:** PostgreSQL via jackc/pgx/v5, sqlc for type-safe queries
- **Money:** shopspring/decimal (arbitrary precision)
- **Auth:** OAuth2 SSO (Google + GitHub), JWT session tokens (golang-jwt/v5)
- **Config:** caarlos0/env/v11 (struct-tag env loading)
- **License:** Apache 2.0

## Build & Run Commands

All commands run from the `backend/` directory:

```bash
# Run the API server
make run                    # or: go run ./cmd/api

# Build binary
make build                  # outputs bin/spend-wise-api

# Run tests
make test                   # or: go test -race -count=1 ./...

# Run a single test
go test -run TestName ./model/

# Vet/lint
go vet ./...
make lint                   # golangci-lint (requires installation)

# Tidy dependencies
make tidy

# Database migrations (requires DATABASE_URL env var)
make migrate-up
make migrate-down

# Generate sqlc code (requires sqlc installation)
make sqlc
```

## Architecture

Pragmatic flat package layout — no `internal/` directory. One-way dependency chain:

```
model ← store ← service ← handler ← cmd/api
                                ↑
                              auth
```

## Project Structure

```
backend/
├── cmd/api/main.go         # Entry point + composition root (run() pattern)
├── model/                  # Entities, value objects, sentinel errors
│   ├── money.go            # Immutable Money value object (shopspring/decimal)
│   ├── daterange.go        # DateRange value object + convenience constructors
│   └── errors.go           # Sentinel domain errors (ErrUserNotFound, etc.)
├── store/                  # PostgreSQL: pgx pool + sqlc generated + queries
│   └── db.go               # Connection pool constructor
├── handler/                # HTTP handlers (Chi), response/request helpers
│   ├── router.go           # Route table + middleware (RequestID, RealIP, Logger, Recoverer)
│   ├── response.go         # respondJSON, respondError helpers
│   └── request.go          # decodeJSON with 1MB limit + unknown field rejection
├── service/                # Business logic (to be populated)
├── auth/                   # OAuth2 providers + JWT middleware (to be populated)
├── config/config.go        # Env-based config with defaults
├── migrations/             # SQL migration files (golang-migrate format)
├── db/queries/             # sqlc SQL source files
├── Makefile                # build, run, test, lint, sqlc, migrate-*, tidy
├── sqlc.yaml               # sqlc config (pgx/v5, NUMERIC→decimal)
├── .golangci.yml           # Linter config
└── .env.example            # All env vars with safe defaults
```

## Key Patterns

- **Value objects** (Money, DateRange): unexported fields, value receivers, immutable
- **Sentinel errors**: use `errors.Is()` for comparison, define in `model/errors.go`
- **Error handling**: wrap and return (`fmt.Errorf("context: %w", err)`), never log AND return
- **HTTP handlers**: always `return` after `respondError`/`respondJSON`
- **Table-driven tests**: all tests use `[]struct{...}` + `t.Run()`
- **Config**: 12-factor style, all from environment variables via struct tags
