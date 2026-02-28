## Why

Developers need a one-command way to start PostgreSQL for local development. Currently, each developer must install and configure PostgreSQL manually, which is error-prone and slows onboarding. A Docker Compose setup provides a reproducible database environment with correct settings, migrations, and credentials out of the box.

## What Changes

- Add `docker-compose.yml` at the project root with a PostgreSQL service (matching `.env.example` defaults: user=postgres, password=postgres, db=spendwise, port=5432)
- Add a `Dockerfile` for the Go API so developers can optionally run the full stack in Docker
- Add a `.env` file (gitignored) workflow — copy from `.env.example`, docker-compose reads it
- Add Makefile targets for common Docker Compose operations (up, down, db-only, logs)
- Update `.gitignore` to exclude `.env` if not already present

## Capabilities

### New Capabilities
- `docker-dev-environment`: Docker Compose configuration for local development — PostgreSQL service with volume persistence, optional Go API service with live rebuild, and Makefile integration

### Modified Capabilities

## Impact

- **New files**: `docker-compose.yml` (project root), `backend/Dockerfile`, updates to `backend/Makefile` and root `.gitignore`
- **No code changes**: Application code is unchanged; only infrastructure/tooling files are added
- **Dependencies**: Requires Docker and Docker Compose installed on the developer machine
- **Existing workflow preserved**: `make run` continues to work for developers who prefer running PostgreSQL natively
