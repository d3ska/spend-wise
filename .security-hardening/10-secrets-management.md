# Secrets Management Review — SpendWise

**Date:** 2026-02-17
**Scope:** Code-level secrets hygiene (local development context)

---

## Overall Assessment: GOOD

The codebase has good secrets hygiene for a local-dev application. No hardcoded real credentials, no secrets in logs or error messages, OAuth secrets require explicit configuration.

---

## Findings & Fixes

| # | Finding | Severity | Status |
|---|---------|----------|--------|
| 1 | No hardcoded real credentials in source | — | **PASS** |
| 2 | No secrets leaked in error messages or logs | — | **PASS** |
| 3 | No debug endpoints exposing config | — | **PASS** |
| 4 | .gitignore missing *.pem/*.key patterns | Medium | **FIXED** |
| 5 | JWT default accepted silently | Low | **FIXED** (warnings added) |
| 6 | DSN() embeds password (could leak in logs) | Low | **FIXED** (RedactedDSN() added) |
| 7 | Test files use clearly fake values | — | **PASS** |
| 8 | .env.example has no real credentials | — | **PASS** |

---

## Detail

### .gitignore Updated
Added patterns: `*.pem`, `*.key`, `*.p12`, `*.pfx`, `.env.*`, `!.env.example`

### Config Validation
- `Load()` now emits `slog.Warn` when JWT secret is the insecure default or < 32 chars
- New `RedactedDSN()` method for safe log output (masks password)
- Doc comment on `DSN()` warning about password inclusion

### Secrets in Memory
Go strings are immutable and cannot be zeroed — this is standard Go behavior and not actionable. All Go JWT libraries and the stdlib TLS implementation use strings for secrets.

### Positive Patterns Found
- OAuth secrets have no defaults (must be explicitly set)
- `CookieSecure` defaults to `true`
- `AUTH_BYPASS` defaults to `false` with loud warning when enabled
- Bank API code logs only value lengths, never the actual values
- `safeRecoverer` prevents stack traces from reaching clients

---

## Files Changed

| File | Change |
|------|--------|
| `.gitignore` | Added private key and env file patterns |
| `config/config.go` | JWT validation warnings, `RedactedDSN()`, doc comments |
| `config/config_test.go` | Test for `RedactedDSN()` |
