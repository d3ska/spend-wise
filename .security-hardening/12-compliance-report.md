# Compliance & Standards Verification — SpendWise

**Date:** 2026-02-17
**Frameworks:** OWASP Top 10 (2021), OWASP ASVS Level 2
**Scope:** Code-level compliance (local development context)

---

## Overall Verdict

| Framework | Rating |
|---|---|
| **OWASP Top 10 (2021)** | **8/10 PASS, 2 PARTIAL** (A02 Crypto, A04 Insecure Design) |
| **OWASP ASVS Level 2** | **PASS with exceptions** — 2 MEDIUM, 4 LOW findings |

---

## OWASP Top 10 (2021) Assessment

| Category | Status | Key Evidence |
|----------|--------|-------------|
| **A01: Broken Access Control** | **PASS** | RBAC enforced via requireMembership, IDOR fixed on all resources, role hierarchy validated |
| **A02: Cryptographic Failures** | **PARTIAL** | JWT HS256 locked, crypto/rand for tokens, cookies secure — but 2 unredacted IBAN log locations |
| **A03: Injection** | **PASS** | sqlc parameterized queries, no raw SQL, sort whitelists, regex validation, JSON auto-escaping |
| **A04: Insecure Design** | **PARTIAL** | Layered arch, RBAC, batch limits — but open redirect in bank redirect_url, invite TOCTOU race |
| **A05: Security Misconfiguration** | **PASS** | Security headers, CORS hardening, unknown fields rejected, panic recovery, server timeouts |
| **A06: Vulnerable Components** | **PASS** | Go modules managed, key deps current |
| **A07: Auth Failures** | **PASS** | OAuth2 SSO, state validation, rate limiting, JWT validation, secure cookies |
| **A08: Data Integrity** | **PASS** | Safe JSON deserialization, unknown fields rejected, batch limits, duplicate detection |
| **A09: Logging Failures** | **PARTIAL** | Structured slog logging, request IDs, generic client errors — but ErrTransactionNotFound missing from error mapper |
| **A10: SSRF** | **PARTIAL** | Single configured API endpoint — but user-supplied redirect_url forwarded unvalidated |

---

## OWASP ASVS Level 2 — Key Requirements

| ASVS Category | Status | Notes |
|---|---|---|
| V2 Authentication | PASS | OAuth2 SSO, no passwords stored |
| V3 Session Management | PASS | JWT 24h, HttpOnly/Secure/SameSite cookies, crypto/rand tokens |
| V4 Access Control | PASS | Server-side RBAC, IDOR prevention, workspace scoping |
| V5 Input Validation | PASS | 1MB body, field lengths, parameterized SQL, output encoding |
| V7 Error/Logging | PARTIAL | 2 IBAN log leaks, 1 missing error mapping |
| V8 Data Protection | PARTIAL | IBAN partially redacted |
| V9 Communications | N/A | Local dev |
| V11 Business Logic | PASS | Rate limiting, batch limits |
| V12 Files/Resources | PASS | Body limits, API response limits |
| V14 Configuration | PASS | Env vars, weak default warnings |

---

## Remaining Findings (consolidated)

### MEDIUM (2)
1. **Unvalidated redirect URL** in bank connection flows — needs domain allowlist
2. **Unvalidated role values** at application layer — DB ENUM catches it, but app-level validation needed

### LOW (4)
3. **Unredacted IBAN** in 2 log locations
4. **Missing ErrTransactionNotFound** in handleServiceError
5. **Invite accept TOCTOU** race condition
6. **No YearMonth format** validation in funding endpoints

### INFO (3)
7. Auth bypass no compile-time guard (acceptable for local dev)
8. SSE event type not defensively sanitized (not exploitable)
9. No pagination on ApplyRules (acceptable for local dev)
