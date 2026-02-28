# Infrastructure Security Controls — SpendWise

**Date:** 2026-02-17
**Scope:** Code-level infrastructure security (local development context)

---

## Summary of Findings & Fixes

| # | Finding | Severity | Status |
|---|---------|----------|--------|
| 1 | SQL injection audit (all queries) | — | **PASS — no injection risks** |
| 2 | SSE connections not cleaned up on shutdown | Medium | **FIXED** |
| 3 | HTTP server missing slowloris protections | Medium | **FIXED** |
| 4 | Panic recovery leaks stack traces to client | Medium | **FIXED** |
| 5 | No per-request context timeout | Medium | **FIXED** |
| 6 | Concurrent access safety audit | — | **PASS — all synchronized** |
| 7 | Error handling consistency audit | — | **PASS — good patterns** |
| 8 | Unbounded bank API response body | Medium | **FIXED** |

---

## Audit Results

### 1. Database Query Safety — PASS
- All sqlc-generated queries use parameterized statements (10 .sql.go files)
- Hand-written queries in `summary_store.go` use `$N` placeholders exclusively
- Dynamic ORDER BY in `transaction_store.go` uses whitelist for columns and strict check for direction
- No SQL injection paths found

### 2. SSE Shutdown Cleanup — FIXED
**Files:** `handler/sse_hub.go`, `handler/handler_sse.go`, `cmd/api/main.go`

Added `Close()` method to `SSEHub` that closes all client channels and clears the map. SSE handler detects closed channels. `sseHub.Close()` called in shutdown sequence.

### 3. Slowloris Protection — FIXED
**File:** `cmd/api/main.go`

Added `ReadHeaderTimeout: 5s` and `IdleTimeout: 120s` to HTTP server config. Prevents slow clients from holding connections open indefinitely.

### 4. Safe Panic Recovery — FIXED
**File:** `handler/router.go`

Replaced Chi's `middleware.Recoverer` (which writes stack traces to response body) with custom `safeRecoverer` that:
- Logs full panic + stack trace server-side via `slog.Error`
- Returns bare 500 status with no body to client
- Preserves `http.ErrAbortHandler` semantics

### 5. Per-Request Context Timeout — FIXED
**File:** `handler/router.go`

Added `middleware.Timeout(30 * time.Second)`. All database operations and the bank API client properly accept and propagate `ctx`. SSE handler already manages its own deadline.

### 6. Concurrent Access Safety — PASS
- SSEHub: `sync.RWMutex`, buffered channels — safe
- StateStore: `sync.Mutex` — safe
- RateLimiter: `sync.Mutex` — safe
- DevBypass: `sync.Once` — safe
- No unprotected shared mutable state found

### 7. Error Handling — PASS
- All store errors consistently wrapped with `fmt.Errorf("context: %w", err)`
- `handleServiceError` maps domain errors to HTTP codes, logs unknown errors server-side
- No silently swallowed errors in handlers
- Transaction rollback errors properly handled

### 8. Bank API Response Body Limit — FIXED
**File:** `banksync/http_client.go`

`io.ReadAll(resp.Body)` replaced with `io.LimitReader(resp.Body, 10MB)`. Prevents OOM from unbounded external API responses.

---

## Graceful Shutdown Sequence (verified)

```
1. Signal received (SIGINT/SIGTERM)
2. cronScheduler.Stop() — waits for running cron jobs
3. sseHub.Close() — closes all SSE connections
4. srv.Shutdown(ctx) — drains in-flight HTTP requests (10s timeout)
5. pool.Close() (defer) — closes database connections
```

## Build/Test Status
- `go build ./...` — PASS
- `go test ./...` — PASS
- `go vet ./...` — PASS
