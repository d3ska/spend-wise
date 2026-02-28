## Context

SpendWise requires PostgreSQL for local development. Currently developers must install PostgreSQL natively and configure it to match `.env.example` defaults (user=postgres, password=postgres, db=spendwise, port=5432). Docker Compose can automate this setup with a single command.

The Go API already reads configuration from environment variables via `config/config.go`. No application code changes are needed — only infrastructure files.

## Goals / Non-Goals

**Goals:**
- One-command PostgreSQL startup for local development (`make docker-db`)
- Optional full-stack mode running both PostgreSQL and the Go API in Docker
- Automatic database creation on first run (via `POSTGRES_DB` env var)
- Volume persistence so data survives container restarts
- Makefile integration for common operations

**Non-Goals:**
- Production Docker deployment (no multi-stage optimized builds, no K8s manifests)
- Automatic migration execution inside Docker (developers run `make migrate-up` manually)
- CI/CD Docker configuration
- Docker Compose profiles or environment-specific overrides

## Decisions

### 1. Single docker-compose.yml with two services

One Compose file at the project root with `db` (always) and `api` (optional) services. Developers use `docker compose up db` for database-only or `docker compose up` for full stack.

**Alternatives considered:**
- Separate `docker-compose.db.yml` and `docker-compose.full.yml` — more files, more confusion, same result
- Only PostgreSQL, no API service — limits usefulness for quick demos

### 2. PostgreSQL 16 with named volume

Use `postgres:16-alpine` for small image size. A named volume (`spendwise-pgdata`) ensures data persists across `docker compose down` and `docker compose up`. Only `docker compose down -v` destroys data.

**Alternatives considered:**
- Bind mount (`./data/postgres`) — pollutes working directory, platform-specific path issues on Windows
- No volume — loses all data on container restart, unusable for development

### 3. Dockerfile uses multi-stage build

Stage 1 (`builder`): `golang:1.26-alpine`, copies source, runs `go build`. Stage 2 (`runner`): `alpine:3.21`, copies binary only. Keeps final image small (~30MB vs ~1GB).

The API service in docker-compose mounts nothing — it builds from the Dockerfile. For live reload during development, developers should run the Go app locally against the Dockerized PostgreSQL.

### 4. Environment variables from .env.example defaults

Docker Compose uses `environment:` directives matching `.env.example` defaults exactly. No `.env` file is required for basic usage. Developers who need custom values can create a `.env` file (already gitignored).

### 5. Makefile targets use `docker compose` (v2 syntax)

Docker Compose v2 uses `docker compose` (space, not hyphen). The old `docker-compose` binary is deprecated. All Makefile targets use the v2 syntax.

## Risks / Trade-offs

- **Port conflict**: If the developer has a native PostgreSQL on port 5432, the Docker container will fail to bind. Mitigation: document in `.env.example` that the port can be changed.
- **Docker required**: Adds Docker as a development dependency. Mitigation: existing `make run` workflow still works for developers with native PostgreSQL.
- **Data loss on `down -v`**: Named volumes are destroyed with `-v` flag. Mitigation: `make docker-down` does NOT use `-v`; a separate `make docker-nuke` target is explicit.
