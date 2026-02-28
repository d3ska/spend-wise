## 1. Docker Compose

- [x] 1.1 Create `docker-compose.yml` at project root — `db` service (postgres:16-alpine, port 5432, named volume `spendwise-pgdata`, env: POSTGRES_USER=postgres, POSTGRES_PASSWORD=postgres, POSTGRES_DB=spendwise, healthcheck), `api` service (build from backend/Dockerfile, port 8080, depends_on db with condition service_healthy, env matching .env.example defaults)

## 2. Dockerfile

- [x] 2.1 Create `backend/Dockerfile` — multi-stage build: stage 1 `golang:1.26-alpine` builder (copy go.mod/go.sum, download deps, copy source, `go build -o /app ./cmd/api`), stage 2 `alpine:3.21` runner (copy binary, expose 8080, entrypoint)

## 3. Makefile Targets

- [x] 3.1 Update `backend/Makefile` — add targets: `docker-db` (docker compose -f ../docker-compose.yml up db -d), `docker-up` (docker compose -f ../docker-compose.yml up -d --build), `docker-down` (docker compose -f ../docker-compose.yml down), `docker-nuke` (docker compose -f ../docker-compose.yml down -v), `docker-logs` (docker compose -f ../docker-compose.yml logs -f)

## 4. Gitignore

- [x] 4.1 Ensure `.env` is in `.gitignore` at project root (add if missing)

## 5. Verification

- [x] 5.1 Run `docker compose config` from project root — must parse without errors
- [x] 5.2 Run `docker compose up db -d` — PostgreSQL must start and accept connections on port 5432
- [x] 5.3 Run `docker compose down` — containers stop, volume preserved
