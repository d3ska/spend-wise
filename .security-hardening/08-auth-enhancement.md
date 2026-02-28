# Authentication & Authorization Enhancement — SpendWise

**Date:** 2026-02-17
**Scope:** Code-level auth improvements (local development context)

---

## Summary of Findings & Fixes

| # | Finding | Severity | Status |
|---|---------|----------|--------|
| 1 | Cross-workspace IDOR in Category operations (UpdateCategory, DeleteCategory) | **HIGH** | **FIXED** |
| 2 | AddMember allows granting Owner role | **MEDIUM** | **FIXED** |
| 3 | Owner can remove themselves, orphaning workspace | **MEDIUM** | **FIXED** |
| 4 | Missing JWT audience claim | Low | **FIXED** |
| 5 | No rate limiting on auth endpoints | Medium | **FIXED** |
| 6 | Auth completeness audit (30+ endpoints) | — | **VERIFIED** |
| 7 | OAuth flow security | — | **VERIFIED SECURE** |
| 8 | Permission escalation paths | — | **AUDITED** |

---

## NEW Bug Fixes

### Fix 1: Category IDOR (HIGH)
**Files:** `service/workspace_service.go`

`UpdateCategory` and `DeleteCategory` checked workspace membership but operated on category IDs without verifying the category belongs to the workspace. An editor in workspace A could modify categories in workspace B.

- `UpdateCategory` now fetches category first and verifies `existing.WorkspaceID != wsID`
- `DeleteCategory` now checks `cat.WorkspaceID != wsID`

### Fix 2: AddMember Role Escalation (MEDIUM)
**File:** `service/workspace_service.go`

`AddMember` had no validation on the role being assigned — an owner could create additional owners. Now rejects `model.RoleOwner` with `ErrInsufficientPermission`.

### Fix 3: Owner Self-Removal (MEDIUM)
**Files:** `service/workspace_service.go`, `model/errors.go`, `handler/handler_workspace.go`

Owner could remove themselves via `DELETE /workspaces/{id}/members/{own-id}`, orphaning the workspace. Now returns `ErrCannotRemoveSelf` when `userID == targetUserID`.

### Fix 4: JWT Audience Claim
**File:** `auth/jwt.go`

Added `aud: "spendwise-api"` claim to token signing and verification for defense-in-depth.

### Fix 5: Auth Rate Limiting
**New file:** `auth/rate_limiter.go`
**Modified:** `handler/router.go`, `cmd/api/main.go`

Per-IP token-bucket rate limiter: 10 req/s, burst 20. Applied to auth routes. Returns 429 with Retry-After header. Stale entries cleaned every 5 minutes.

---

## Authorization Completeness Audit

Full endpoint-by-endpoint audit verified:
- **30+ endpoints** checked for workspace membership, role level, and IDOR protection
- All mutating operations require `editor+` role
- All read operations require `viewer+` role
- Workspace ownership verified for all resource-scoped operations
- Self-role-change prevention confirmed
- Invite role restriction confirmed (editor/viewer only)

## OAuth Flow Verification

- Code exchange: server-side only, client never sees OAuth tokens
- State parameter: validated, single-use, 10-min TTL
- GitHub email: requires Primary + Verified
- HTTP timeouts: 10s on all OAuth client calls
- User linking: email-based, overwrites provider (acceptable for local dev)

## Remaining Recommendations (no code changes needed for local dev)

1. Server-side token blocklist for logout
2. Shorter JWT expiry with refresh tokens
3. Multi-provider user linking via join table
4. Rate limiting on invite code endpoint
5. SSE re-validates membership on remove events
