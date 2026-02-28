# Critical Code-Level Fixes — SpendWise

**Date:** 2026-02-17
**Scope:** Code logic and security bugs only (local development context)
**Fixes Applied:** 6

---

## Summary

| # | Vulnerability | Severity | Files Changed | Status |
|---|--------------|----------|---------------|--------|
| 1 | Cross-Workspace IDOR | Critical | `service/transaction_service.go`, `service/rule_service.go`, `store/rule_store.go`, `store/rules.sql.go`, `db/queries/rules.sql` | **FIXED** |
| 2 | OAuth State Not Validated | High | `handler/handler_auth.go`, `auth/state_store.go` (NEW), `cmd/api/main.go` | **FIXED** |
| 3 | Regex Validation Missing | Medium | `service/rule_service.go`, `model/errors.go`, `handler/handler_workspace.go`, `store/rule_store.go` | **FIXED** |
| 4 | Sensitive Bank Data in Logs | Medium | `banksync/http_client.go`, `service/bank_connection_service.go` | **FIXED** |
| 5 | SSE Connection Exhaustion | Medium | `handler/sse_hub.go`, `handler/handler_sse.go` | **FIXED** |
| 6 | Import Batch Size Limit | Low-Medium | `handler/handler_transaction.go` | **FIXED** |

---

## Fix 1: Cross-Workspace IDOR (OWASP A01:2021 — Broken Access Control)

**The Bug:** Services check workspace membership but fetch/modify resources by ID alone without verifying the resource belongs to the requested workspace. An attacker member of Workspace A could read/modify/delete resources from Workspace B.

**Changes:**
- `TransactionService.GetTransaction()` — now fetches transaction then verifies `tx.WorkspaceID == wsID`
- `TransactionService.DeleteTransaction()` — same pattern: fetch first, verify workspace, then delete
- `RuleService.Update()`, `Delete()`, `Toggle()` — new `requireRuleInWorkspace()` helper verifies `rule.WorkspaceID == wsID`
- New `RuleStore.GetByID()` method + SQL query added to support rule lookup
- New `ErrRuleNotFound` sentinel error

## Fix 2: OAuth State Parameter Validation (OWASP A01:2021 — CSRF)

**The Bug:** `GetAuthURL()` generates a state parameter but `HandleCallback()` never validates it, enabling CSRF attacks on the OAuth flow.

**Changes:**
- New `auth/state_store.go` — in-memory state store with TTL (10 min) and single-use consumption
- `GetAuthURL()` — stores state in state store before returning
- `HandleCallback()` — validates state against store before code exchange; rejects invalid/expired states
- `cmd/api/main.go` — initializes StateStore and passes to AuthHandler

## Fix 3: Regex Validation at Rule Creation (OWASP A03:2021 — Injection)

**The Bug:** User-supplied regex patterns accepted without validation. Invalid patterns silently skipped at resolution time.

**Changes:**
- New `validatePattern()` in `rule_service.go` — enforces 500-char max length + `regexp.Compile()` check
- Applied to both `Create()` and `Update()` methods
- New sentinel errors: `ErrRuleInvalidPattern`, `ErrRulePatternTooLong`
- `handleServiceError` updated to return 400 Bad Request for these errors
- Silent `continue` in `rule_store.go` replaced with `slog.Warn` for observability

## Fix 4: Sensitive Bank Data Redacted from Logs (OWASP A09:2021)

**The Bug:** Bank API HTTP client logs full request/response bodies (IBANs, session IDs, account UIDs) at default level using `log.Printf`.

**Changes:**
- `banksync/http_client.go` — replaced `log.Printf` with `slog.Debug`, logs byte counts instead of raw bodies
- `bank_connection_service.go` — new `redactIBAN()` helper masks IBANs to `XX****1234` format
- Removed session IDs and external UIDs from INFO-level logs

## Fix 5: SSE Connection Limits

**The Bug:** No limit on SSE connections per workspace — resource exhaustion possible.

**Changes:**
- `sse_hub.go` — `Subscribe()` now returns error, enforces `maxConnectionsPerWorkspace = 50`
- `handler_sse.go` — handles limit error with 429 Too Many Requests

## Fix 6: Import Batch Size Cap

**The Bug:** `ImportTransactions` accepts unbounded JSON arrays.

**Changes:**
- `handler_transaction.go` — added `maxImportBatchSize = 500` check after JSON decode

---

## Test Updates

- `service/rule_service_test.go` — updated mock for new `GetByID` interface
- `service/transaction_service_test.go` — fixed test data for workspace ownership checks
- `handler/handler_transaction_test.go` — updated mock + test data for new behavior

---

## Regression Test Requirements

| Fix | Test Scenario |
|-----|--------------|
| IDOR | Attempt to GET/DELETE a transaction belonging to workspace B while authenticated to workspace A → expect 404 |
| IDOR | Attempt to UPDATE/DELETE/TOGGLE a rule belonging to workspace B → expect 404 |
| OAuth State | Callback with empty state → expect 400 |
| OAuth State | Callback with invalid/expired state → expect 400 |
| OAuth State | Callback with valid state succeeds; reuse same state → expect 400 (single-use) |
| Regex | Create rule with invalid regex → expect 400 |
| Regex | Create rule with pattern > 500 chars → expect 400 |
| SSE | Open 51 SSE connections to same workspace → expect 429 on 51st |
| Import | Send array of 501 transactions → expect 400 |
