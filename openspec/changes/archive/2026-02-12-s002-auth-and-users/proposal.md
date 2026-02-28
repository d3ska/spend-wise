## Why

Users need to authenticate before accessing any SpendWise functionality. For MVP, OAuth2 SSO (Google + GitHub) eliminates password storage liability, removes the need for password reset flows, and provides lower registration friction with verified emails and profile data for free. This is the first feature change — without auth, no protected endpoint can exist.

**Depends on:** `s001-project-foundation` (project structure, config, HTTP server, domain value objects)

## What Changes

- Create the `User` entity in `model/user.go` with `UserID` typed wrapper, `AuthProvider` type (google|github), and fields: id, email (unique), display_name, avatar_url, provider, provider_id — no password fields
- Create database migration `000001_create_users` with `auth_provider` PostgreSQL enum, email unique index
- Create sqlc queries for users (`db/queries/users.sql`)
- Create `store/user_store.go` implementing user persistence
- Create `auth/provider.go` with `Provider` interface (AuthCodeURL, Exchange, FetchUser -> UserInfo)
- Create `auth/google.go` — Google OAuth2 config, token exchange, profile fetch from Google userinfo API
- Create `auth/github.go` — GitHub OAuth2 config, token exchange, profile fetch from GitHub user API
- Create `auth/jwt.go` — JWT signing/verification helpers
- Create `auth/middleware.go` — JWT Bearer validation middleware, extracts `UserID` into request context
- Create `service/auth_service.go` with single method: `Authenticate(ctx, provider, code) -> (token, user, error)` — exchanges OAuth code, fetches profile, upserts user by email, issues JWT
- Create `handler/handler_auth.go` with endpoints:
  - `GET /api/v1/auth/{provider}/url` — returns OAuth authorization URL with CSRF state
  - `POST /api/v1/auth/{provider}/callback` — exchanges code for JWT + user profile
  - `GET /api/v1/auth/me` — returns current authenticated user (protected)
- Wire auth service and handler into `cmd/api/main.go` router
- OAuth HTTP client uses explicit 10s timeout (100go.co #81), response bodies always closed (#79)

## Capabilities

### New Capabilities
- `oauth-auth`: OAuth2 SSO authentication with Google and GitHub providers, JWT session tokens, user creation-on-first-login, CSRF state protection, provider-agnostic interface for adding future providers
- `user-management`: User entity with OAuth provider identity, user store, create-or-find-by-email semantics for cross-provider account linking

### Modified Capabilities
- `http-server`: Add JWT auth middleware and auth routes to the Chi router

## Impact

- **API:** 3 new endpoints (`/auth/{provider}/url`, `/auth/{provider}/callback`, `/auth/me`). All subsequent changes depend on the JWT middleware from this change.
- **Database:** Migration 001 creates `users` table with `auth_provider` enum
- **Dependencies:** `golang.org/x/oauth2` (already in go.mod from foundation)
- **Config:** Env vars already defined: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_REDIRECT_URL`, `JWT_SECRET`, `JWT_EXPIRY`
- **External services:** Requires Google Cloud Console OAuth credentials and GitHub OAuth App credentials for testing
