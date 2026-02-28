# Security Monitoring & Observability — SpendWise

**Date:** 2026-02-17
**Scope:** Code-level logging and observability improvements (local development context)

---

## Current State

The application uses Go's `log/slog` for structured logging with:
- Request ID middleware for correlation
- Authentication failure logging
- Panic recovery with server-side stack traces
- Bank API PII redaction (partial)
- Generic error responses to clients

---

## Security Event Logging Audit

### Currently Logged

| Event | File | Level |
|---|---|---|
| Authentication failure | `handler/handler_auth.go:98` | Error |
| Panic with stack trace | `handler/router.go:189` | Error |
| Unhandled service errors | `handler/handler_workspace.go:280` | Error |
| Bank API HTTP errors | `banksync/http_client.go:169` | Warn |
| Bank connection ownership mismatch | `service/bank_connection_service.go:137` | Warn |
| CORS wildcard rejection | `handler/router.go:216` | Warn |
| Weak JWT secret | `config/config.go:101-104` | Warn |

### NOT Logged (Gaps)

| Event | File | Priority |
|---|---|---|
| **Successful logins** | `handler/handler_auth.go:103` | **P1** |
| **Rate limit triggers** | `auth/rate_limiter.go:90` | **P1** |
| **JWT validation failures** | `auth/middleware.go:19-26` | **P1** |
| **Authorization denials** | `handler/handler_workspace.go:259-262` | **P1** |
| Member additions | `service/invite_service.go:129`, `workspace_service.go` | P2 |
| Member removals | `handler/handler_workspace.go:216` | P2 |
| Role changes | `service/workspace_service.go:176` | P2 |
| Workspace deletion | `service/workspace_service.go:151` | P2 |
| Bank connection deletion | `service/bank_connection_service.go:233` | P3 |
| Bank connection state changes | `service/bank_connection_service.go:238-303` | P3 |

---

## PII Redaction Gaps

| Location | Issue |
|----------|-------|
| `service/bank_sync_service.go:139` | Raw IBAN logged — should use `redactIBAN()` |
| `service/bank_account_linking_service.go:76` | Raw IBAN logged — should use `redactIBAN()` |

The `redactIBAN()` function exists in `bank_connection_service.go:18-23` but is not shared.

---

## Log Structure Issues

1. **Request ID not propagated to service layer** — service-layer logs can't be correlated to HTTP requests
2. **User ID missing from security events** — `handleServiceError` doesn't have access to caller identity
3. **No correlation IDs for bank sync** — multi-step sync produces many log lines with no grouping

---

## Prioritized Recommendations

### P1 — High-Value, Low-Effort (1 line each)

1. **Log successful logins** at `handler/handler_auth.go` after line 103:
   ```go
   slog.Info("user authenticated", "user_id", user.ID, "email", user.Email, "provider", providerName)
   ```

2. **Log rate limit triggers** at `auth/rate_limiter.go` inside the 429 branch:
   ```go
   slog.Warn("rate limit exceeded", "ip", ip, "path", r.URL.Path)
   ```

3. **Log JWT validation failures** at `auth/middleware.go` before 401 returns:
   ```go
   slog.Warn("jwt validation failed", "error", err, "path", r.URL.Path, "ip", r.RemoteAddr)
   ```

4. **Log authorization denials** at `handler/handler_workspace.go` in the Forbidden cases:
   ```go
   slog.Warn("authorization denied", "error", err)
   ```

### P2 — Audit Trail for Access Control Mutations

5. **Log member additions** (workspace_service.go, invite_service.go)
6. **Log member removals** (workspace_service.go)
7. **Log role changes** (workspace_service.go) — who changed whose role to what
8. **Log workspace deletion** (workspace_service.go)

### P3 — Structural Improvements

9. **Fix 2 IBAN PII leaks** — export `redactIBAN` and use in bank_sync_service + bank_account_linking_service
10. **Propagate request_id to service-layer logs** — context-aware slog middleware
11. **Add bank sync correlation ID** — UUID generated at SyncAll start, threaded through child calls
