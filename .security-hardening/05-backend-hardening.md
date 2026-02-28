# Backend Security Hardening — SpendWise

**Date:** 2026-02-17
**Scope:** Code-level backend hardening (local development context)
**Changes Applied:** 6 categories

---

## Summary

| # | Hardening Measure | Files Changed | Status |
|---|------------------|---------------|--------|
| 1 | Security Headers Middleware | `handler/router.go` | **APPLIED** |
| 2 | CORS Hardening | `handler/router.go` | **APPLIED** |
| 3 | Error Response Sanitization | `handler/handler_workspace.go` | **APPLIED** |
| 4 | JWT Claims Enrichment | `auth/jwt.go` | **APPLIED** |
| 5 | Request Validation Hardening | `model/workspace.go`, `model/category.go`, `model/transaction.go`, `model/errors.go`, `service/workspace_service.go`, `handler/handler_workspace.go` | **APPLIED** |
| 6 | URL Parameter Encoding | `banksync/http_client.go` | **APPLIED** |

---

## 1. Security Headers Middleware

**File:** `handler/router.go`

Added `securityHeaders` middleware as the first middleware in the Chi router chain:
- `X-Content-Type-Options: nosniff` — prevents MIME sniffing
- `X-Frame-Options: DENY` — prevents clickjacking
- `Referrer-Policy: strict-origin-when-cross-origin` — limits referrer leakage
- `X-XSS-Protection: 0` — disables legacy XSS Auditor (rely on CSP)
- `Permissions-Policy: camera=(), microphone=(), geolocation=()` — restricts browser features

## 2. CORS Hardening

**File:** `handler/router.go`

`parseOrigins()` now explicitly rejects wildcard (`"*"`) origins and logs a warning. This prevents accidental misconfiguration where `AllowCredentials: true` + wildcard would allow any site to make credentialed requests.

## 3. Error Response Sanitization

**File:** `handler/handler_workspace.go`

The `default` case in `handleServiceError` now logs the actual error via `slog.Error` before returning the generic `"internal server error"` to clients. Operators get diagnostic info; clients never see internal details.

## 4. JWT Claims Enrichment

**File:** `auth/jwt.go`

- `SignToken` now includes `Issuer: "spendwise"` and `NotBefore: now` claims
- `VerifyToken` now enforces issuer via `jwt.WithIssuer("spendwise")`
- Prevents cross-service token reuse and pre-dated token attacks

## 5. Request Validation Hardening

### 5a. Workspace name/description limits
- `model/workspace.go` — `MaxWorkspaceNameLen = 100`, `MaxWorkspaceDescriptionLen = 500`, `Validate()` method
- New errors: `ErrWorkspaceNameTooLong`, `ErrWorkspaceDescriptionTooLong`
- Wired into `handleServiceError` as 400 Bad Request

### 5b. Category name validation
- `model/category.go` — `MaxCategoryNameLen = 100`, `Validate()` method
- New errors: `ErrCategoryNameRequired`, `ErrCategoryNameTooLong`
- Added `cat.Validate()` in `CreateCategory` and `UpdateCategory`

### 5c. Transaction amount and date range validation
- `model/transaction.go` — `ValidateEntries()` now checks:
  - Amount absolute value ≤ 999,999,999
  - Date between 2000-01-01 and 1 year in the future
- New errors: `ErrTransactionAmountTooLarge`, `ErrTransactionDateTooFarInPast`, `ErrTransactionDateTooFarInFuture`
- Tests updated for new validation rules

## 6. URL Parameter Encoding

**File:** `banksync/http_client.go`

- `country` parameter: `url.QueryEscape()`
- `accountUID`: `url.PathEscape()` (path segment)
- `date_from`, `date_to`, `continuation_key`: `url.QueryEscape()`
- Prevents query parameter injection and path traversal against Enable Banking API

---

## All Files Modified

| File | Changes |
|------|---------|
| `handler/router.go` | Security headers middleware + CORS wildcard rejection |
| `handler/handler_workspace.go` | Error logging in default case + new validation errors |
| `auth/jwt.go` | Added iss/nbf claims, enforce issuer on verify |
| `model/workspace.go` | Validation method with length limits |
| `model/category.go` | Validation method with length limits |
| `model/transaction.go` | Amount ceiling + date range validation |
| `model/errors.go` | 5 new sentinel errors |
| `model/transaction_test.go` | Updated for new validation |
| `service/workspace_service.go` | Category validation calls |
| `banksync/http_client.go` | URL encoding for all parameters |

---

## Build/Test Status
- `go build ./...` — PASS
- `go test ./...` — PASS
- `go vet ./...` — PASS
