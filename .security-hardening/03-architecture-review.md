# Architecture Security Review — SpendWise

**Date:** 2026-02-17
**Reviewer:** Security Architecture Agent
**Target:** SpendWise collaborative household financial ledger
**Scope:** Full codebase review — all Go source files, migrations, config, git history

---

## Executive Summary

After thorough review of every source file in the SpendWise backend codebase, the assessment identified systemic architectural security gaps that collectively represent serious risk for a financial application handling bank account connections, IBANs, and transaction data. The codebase has solid foundational patterns (parameterized queries via sqlc, role-based access control, HttpOnly cookies) but has critical gaps in secret management, authentication bypass guardrails, input sanitization, cross-workspace authorization, and observability.

---

## 1. Security Architecture Assessment

### 1.1 Authentication & Session Management

**JWT Implementation** (`auth/jwt.go`)
- Uses HS256 with `jwt.WithValidMethods([]string{"HS256"})` — prevents algorithm confusion attacks
- **VULN-002 confirmed**: JWT secret defaults to `"change-me-in-production"` in `config/config.go:59` with no runtime validation
- **Missing JWT claims**: No `iss`, `aud`, `nbf`, `jti` — enables cross-service token reuse and replay
- **No token revocation**: Logout handler (`handler_auth.go:102-108`) only clears client cookie — JWT remains valid until 24h expiry
- No refresh token rotation mechanism

### 1.2 OAuth2 Flow

**VULN-004 confirmed**: `GetAuthURL()` generates cryptographic state parameter but `HandleCallback()` never validates it:
```go
// handler_auth.go — state is decoded but never compared to anything
var body struct {
    Code  string `json:"code"`
    State string `json:"state"`  // UNUSED
}
```
This is a textbook OAuth CSRF vulnerability.

### 1.3 Dev Authentication Bypass

**VULN-003 confirmed**: `AUTH_BYPASS=true` env var activates `DevBypassMiddleware` with no build tag, no `APP_ENV` check, no compile-time guard. A single environment variable misconfiguration in production bypasses all authentication.

### 1.4 Authorization Architecture — Cross-Workspace IDOR Vulnerabilities

The RBAC implementation via `WorkspaceService.requireMembership()` is well-structured, **but multiple services have IDOR vulnerabilities**:

| Service | Method | File:Line | Issue |
|---------|--------|-----------|-------|
| TransactionService | `GetTransaction()` | `transaction_service.go:200-205` | Checks workspace membership but fetches by `txID` alone — no workspace scoping |
| TransactionService | `DeleteTransaction()` | `transaction_service.go:225-230` | Same — deletes by txID without verifying workspace ownership |
| RuleService | `Delete()` | `rule_service.go:81-91` | Checks workspace membership but deletes by `ruleID` alone |
| RuleService | `Update()` | `rule_service.go:68-78` | Same pattern |
| RuleService | `Toggle()` | `rule_service.go` | Same pattern |

**Impact**: Any authenticated user who is a member of workspace A can read/modify/delete transactions and rules from workspace B if they know the resource ID.

### 1.5 Bank API Security

- **VULN-001 confirmed**: RSA private key (`de06bd2e-afc5-48a7-83e3-3c49660f123f.pem`) committed to git. `.gitignore` missing `*.pem` patterns.
- **VULN-006 confirmed**: `banksync/http_client.go` logs complete request/response bodies at default level via `log.Printf` — exposes IBANs, session IDs, account UIDs. Additionally uses stdlib `log.Printf` instead of `slog`, bypassing centralized log management.
- `BankConnectionService.Complete()` also logs sensitive data (IBANs, session IDs, account UIDs) at INFO level.
- URL path injection risk: path parameters (`country`, `accountUID`) not URL-encoded in Enable Banking API calls.

### 1.6 Input Validation

- **VULN-005 confirmed**: User-supplied regex patterns compiled on every resolution call with no validation at creation time. Invalid patterns silently skipped.
- **SQL injection protection**: Strong — sqlc parameterized queries throughout. Dynamic sort column uses whitelist. Sort direction restricted to hard-coded "ASC"/"DESC".
- **Request body limits**: `decodeJSON()` enforces 1MB via `http.MaxBytesReader` — good.
- **Import endpoint**: Accepts unbounded array `[]importTransactionRequest` — no batch size limit.

### 1.7 Network & Transport Security

- **HTTP server** (`cmd/api/main.go:150-155`): Uses `ListenAndServe` — **plaintext HTTP, no TLS termination**
- **Database connections**: Default `sslmode=disable` — all financial data unencrypted between app and database
- **Enable Banking API**: Uses system default TLS — no certificate pinning, no TLS version enforcement
- **No reverse proxy documented**: No nginx/Envoy/cloud load balancer configuration

### 1.8 Cookie Security

Cookie configuration is well-implemented:
- `HttpOnly: true` — prevents JavaScript access
- `SameSite: Lax` — CSRF protection
- `Secure` configurable, defaults to `true`
- **However**: `.env.example` has `COOKIE_SECURE=false` — developers copying this will deploy with insecure cookies

### 1.9 Error Handling & Information Leakage

- Generic error responses on `default` case — good
- Some domain errors returned directly via `err.Error()` — could leak internal state
- Error messages like "not a workspace member" usable for workspace enumeration

### 1.10 SSE Resource Exhaustion

- `SSEHub` has no limit on connections per workspace or per user
- `StreamEvents` handler disables write deadline: `rc.SetWriteDeadline(time.Time{})`
- Attacker workspace member could open thousands of SSE connections, exhausting server resources

---

## 2. Zero-Trust Architecture Recommendations

### 2.1 Config Hardening — Never Trust the Default

Add `Validate()` method to `Config` struct that fails startup on:
- JWT secret is default or < 32 bytes
- AUTH_BYPASS enabled when `APP_ENV=production`
- DB_SSLMODE is `disable` when `APP_ENV=production`
- Enable Banking key file permissions are world-readable

### 2.2 Verify Every Resource Belongs to the Requester

Every data access must verify workspace scoping:
```go
// After GetByID, verify workspace ownership
tx, err := s.txStore.GetByID(ctx, txID)
if tx.WorkspaceID != wsID {
    return model.ErrTransactionNotFound
}
```
Apply to: `TransactionService.GetTransaction()`, `DeleteTransaction()`, `RuleService.Update()`, `Delete()`, `Toggle()`, and all category operations.

### 2.3 OAuth State Verification

Implement HMAC-signed state tokens:
- Sign state with HMAC-SHA256 using JWT secret before sending to client
- Verify HMAC signature on callback before code exchange
- Add state expiry (5 minutes)

### 2.4 Compile-Time Dev Bypass Guard

Use Go build tags (`//go:build dev`) to exclude dev bypass from production binaries entirely. Add `APP_ENV` runtime check as defense-in-depth.

### 2.5 Least-Privilege Database Access

- Create separate PostgreSQL roles: `spendwise_readonly`, `spendwise_readwrite`, `spendwise_admin` (migrations only)
- Enforce `sslmode=verify-full` with CA certificate in production
- Use pgx TLS configuration to enforce TLS 1.3

### 2.6 Bank API Client Hardening

- Load RSA key from secrets manager, never from disk in production
- Add `*.pem` and `*.key` to `.gitignore`
- URL-encode all path parameters in Enable Banking API calls
- Pin TLS 1.3 for banking API connections
- Replace `log.Printf` with `slog` and redact all PII

---

## 3. Service Mesh Security Requirements

### 3.1 Current Monolith Network Controls

| Control | Current State | Requirement |
|---------|--------------|-------------|
| TLS termination | Plain HTTP `ListenAndServe` | TLS 1.3 or reverse proxy with TLS |
| Database TLS | `sslmode=disable` default | `sslmode=verify-full` with CA cert |
| Enable Banking TLS | System default | Pin TLS 1.3, explicit root CA |
| Health check | `/healthz` exposed publicly | Bind to internal interface or protect |

### 3.2 Future Service Boundaries

```
[API Gateway / Envoy]
  ├── Auth Service (auth/ + service/auth_service.go)
  ├── Workspace Service (service/workspace_service.go, invite_service.go)
  ├── Transaction Service (service/transaction_service.go)
  ├── Bank Sync Service (banksync/ + service/bank_*_service.go)
  └── SSE/Notification Service (handler/sse_hub.go)
```

All inter-service communication: mTLS with short-lived certificates (SPIFFE/SPIRE).

### 3.3 Secrets Inventory

| Secret | Current Location | Recommended |
|--------|-----------------|-------------|
| JWT signing key | Env var with weak default | Secrets manager, 30-day rotation |
| Enable Banking RSA key | File on disk, committed to git | Secrets manager, never on disk |
| OAuth client secrets | Env vars | Secrets manager with audit logging |
| Database password | Env var with default `postgres` | Secrets manager, short-lived IAM creds |

### 3.4 Observability Gaps

| Security Event | Current State | Required |
|----------------|---------------|----------|
| Auth success/failure | Partial logging | Full structured logging with user ID, IP |
| Authorization denial | Not logged | Log every denial with context |
| Bank connection ops | Logged with sensitive data | Log with PII redaction |
| Token issuance | Not logged | Log all token operations |
| Dev bypass activation | Logged as warning | Critical alert |
| Rule creation/modification | Not logged | Audit trail required |
| Workspace member changes | Not logged | Audit trail required |

---

## 4. Data Classification Matrix

### Classification Levels

| Level | Definition | Encryption at Rest | Transit | Retention | Access Logging |
|-------|-----------|-------------------|---------|-----------|---------------|
| **RESTRICTED** | Direct financial harm or regulatory violation | AES-256, column-level | TLS 1.3 mandatory | Per regulation, then crypto-shred | Every access |
| **CONFIDENTIAL** | Personal/financial, privacy-regulated | AES-256, database-level | TLS 1.3 mandatory | 7 years (financial) | Mutations logged |
| **INTERNAL** | Business data, no regulatory impact | Database-level | TLS 1.2+ | 3 years | Aggregate metrics |
| **PUBLIC** | Non-sensitive, intentionally shared | None | TLS recommended | Indefinite | None |

### SpendWise Data Elements

| Data Element | Classification | Current Protection | Gap |
|-------------|---------------|-------------------|-----|
| Enable Banking RSA key | RESTRICTED | **None — committed to git** | CRITICAL: Remove, rotate, secrets manager |
| JWT signing secret | RESTRICTED | Env var with weak default | CRITICAL: Enforce strong secret |
| Bank session IDs | RESTRICTED | Parameterized queries | No column-level encryption, logged at INFO |
| IBANs | CONFIDENTIAL | Parameterized queries | No encryption at rest, logged at INFO, in API responses |
| Bank account UIDs | CONFIDENTIAL | Parameterized queries | Logged at INFO level |
| Transaction amounts | CONFIDENTIAL | Stored as NUMERIC | No column-level encryption |
| Transaction descriptions | CONFIDENTIAL | Parameterized queries | Contains merchant names, payment refs |
| User emails | CONFIDENTIAL | Database-level only | Timing attack on unique index |
| OAuth client secrets | RESTRICTED | Env vars, no defaults | Needs secrets manager |
| Database password | RESTRICTED | Env var, default `postgres` | Must not use default |
| Invite codes | CONFIDENTIAL | 128-bit random, hex-encoded | Time-limited, single-use — good |
| User display names | INTERNAL | None specific | Exposed in invite preview |
| Workspace names | INTERNAL | Membership checks | Exposed in invite preview |
| Category names | INTERNAL | Membership checks | Adequate |
| Rule patterns | INTERNAL | Membership checks | Needs validation |

### Data Flow Diagram — Security Gaps

```
[Browser] --HTTP (no TLS!)--> [Go HTTP Server] --Plaintext (sslmode=disable)--> [PostgreSQL]
                                    |
                                    |--HTTPS--> [Google OAuth API]
                                    |--HTTPS--> [GitHub OAuth API]
                                    |--HTTPS--> [Enable Banking API]
                                              (RSA-signed JWT auth)
                                              (full response bodies logged)
```

**Critical gaps**: No TLS on app server, no TLS on database connection, PII logged in bank API responses.

---

## 5. Architecture Security Score

| Category | Score | Notes |
|----------|-------|-------|
| Authentication | 3/10 | JWT forgery via default secret, no token revocation, no MFA |
| Authorization | 5/10 | Good RBAC pattern but IDOR vulnerabilities in multiple services |
| Data Protection | 3/10 | No encryption at rest/transit, PII in logs, key in git |
| Input Validation | 6/10 | Strong SQL injection prevention, weak regex/batch validation |
| Network Security | 2/10 | No TLS, no DB encryption, no rate limiting |
| Secrets Management | 1/10 | RSA key in git, weak defaults, no secrets manager |
| Observability | 3/10 | Partial logging, no audit trail, sensitive data logged |
| Supply Chain | 5/10 | Vulnerable dependencies, no SBOM, no scanning in CI |

**Overall Architecture Security Score: 3.5/10**

---

## 6. Key Architecture Concerns Summary

1. **Cross-workspace IDOR** (NEW finding, not in vulnerability scan) — multiple services verify workspace membership but don't verify the resource belongs to the workspace
2. **No transport encryption** — HTTP server and database connections are plaintext
3. **Secrets anti-pattern** — critical secrets (RSA key, JWT secret, DB password) all have insecure defaults or are committed to source control
4. **Logging exposes PII** — bank account data (IBANs, session IDs, UIDs) logged at INFO level
5. **No defense-in-depth** — single points of failure in auth (one JWT secret, one bypass flag)
6. **No token lifecycle management** — no revocation, no rotation, no short-lived tokens
