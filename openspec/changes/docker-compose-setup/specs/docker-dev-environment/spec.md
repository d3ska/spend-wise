## ADDED Requirements

### Requirement: PostgreSQL service via Docker Compose
The system SHALL provide a `docker-compose.yml` at the project root that includes a PostgreSQL 16 service (`db`) with credentials matching `.env.example` defaults (user=postgres, password=postgres, db=spendwise, port=5432). Data SHALL be persisted using a named Docker volume.

#### Scenario: Start PostgreSQL only
- **WHEN** developer runs `docker compose up db -d`
- **THEN** a PostgreSQL 16 container starts on port 5432 with database `spendwise` created automatically

#### Scenario: Data persists across restarts
- **WHEN** developer runs `docker compose down` followed by `docker compose up db -d`
- **THEN** all previously inserted data is still present in the database

#### Scenario: Data destroyed on volume removal
- **WHEN** developer runs `docker compose down -v`
- **THEN** the named volume is removed and all data is lost

### Requirement: Go API service via Docker Compose
The system SHALL include an `api` service in `docker-compose.yml` that builds the Go application from a `Dockerfile` in `backend/`. The API service SHALL depend on the `db` service and connect using the same default credentials.

#### Scenario: Start full stack
- **WHEN** developer runs `docker compose up -d`
- **THEN** both PostgreSQL and the Go API start, with the API listening on port 8080

#### Scenario: API waits for database
- **WHEN** docker compose starts both services
- **THEN** the API service SHALL depend on the `db` service and use a health check to wait for PostgreSQL readiness

### Requirement: Dockerfile for Go API
The system SHALL provide a `backend/Dockerfile` using a multi-stage build: a builder stage with `golang:1.26-alpine` that compiles the binary, and a runner stage with `alpine:3.21` that contains only the compiled binary.

#### Scenario: Build produces minimal image
- **WHEN** the Dockerfile is built
- **THEN** the final image contains only the compiled Go binary and Alpine base (no Go toolchain, no source code)

### Requirement: Makefile targets for Docker operations
The `backend/Makefile` SHALL include targets for common Docker Compose operations that delegate to the project root `docker-compose.yml`.

#### Scenario: Database-only mode
- **WHEN** developer runs `make docker-db`
- **THEN** only the PostgreSQL container starts in detached mode

#### Scenario: Full stack mode
- **WHEN** developer runs `make docker-up`
- **THEN** both PostgreSQL and API containers start in detached mode

#### Scenario: Stop containers
- **WHEN** developer runs `make docker-down`
- **THEN** all containers stop and are removed, but volumes are preserved

#### Scenario: Destroy everything
- **WHEN** developer runs `make docker-nuke`
- **THEN** all containers stop and volumes are removed (complete reset)
