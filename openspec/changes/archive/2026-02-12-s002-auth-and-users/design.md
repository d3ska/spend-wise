## Context

The project foundation (`s001-project-foundation`) established the flat package layout (`model/`, `store/`, `handler/`, `service/`, `auth/`, `config/`), Chi router with healthz, JSON helpers, pgx connection pool, and configuration loading. Auth config (JWT secret/expiry, OAuth client credentials) is already defined in `config/config.go`.

This change adds the first real feature: OAuth2 SSO authentication. Every subsequent change (workspaces, transactions, etc.) depends on the JWT middleware and user identity established here.

## Goals / Non-Goals

**Goals:**
- OAuth2 authentication via Google and GitHub — no passwords
- JWT session tokens for stateless API authentication
- User entity with create-on-first-login and cross-provider account linking by email
- Provider-agnostic interface so adding Apple/Microsoft later is one file + one map entry
- JWT middleware that injects `UserID` into request context for all protected routes

**Non-Goals:**
- No refresh tokens — JWT is short-lived (24h default), user re-authenticates after expiry
- No email verification — OAuth providers already verify emails
- No user profile editing — display_name and avatar come from the provider
- No user deletion or account management
- No rate limiting on auth endpoints (future concern)

## Decisions

### 1. Provider interface in `auth/` package

The `auth/` package defines a `Provider` interface with three methods: `AuthCodeURL(state)`, `Exchange(ctx, code)`, `FetchUser(ctx, token)`. Each provider (Google, GitHub) is a separate file implementing this interface.

The `AuthService` receives a `map[model.AuthProvider]auth.Provider` — adding a provider means one new file and one map entry in `main.go`.

**Alternative considered:** Separate `oauth/` package. Rejected — the `auth/` package already exists for JWT middleware. Keeping OAuth providers alongside JWT helpers avoids a package that would have only 2-3 files.

### 2. JWT with HS256, not RS256

JWT tokens are signed with HMAC-SHA256 using a shared secret from config. HS256 is simpler (one secret vs. key pair management) and sufficient when the same backend both signs and verifies.

**Alternative considered:** RS256 (asymmetric). Better for microservices where different services verify tokens without the signing key. Overkill for a monolith.

### 3. User upsert by email for cross-provider linking

When a user authenticates, the service looks up by email first. If found, updates provider info and returns the existing user. If not found, creates a new user. This means a user who first logged in via Google can later log in via GitHub (if same email) and get the same account.

**Alternative considered:** Separate accounts per provider. Simpler but confusing — same person would have two identities and separate workspaces.

### 4. CSRF state via secure random token

The OAuth `state` parameter uses a 32-byte random token (base64url-encoded). The frontend stores it before redirect and sends it back for validation. This prevents CSRF attacks on the callback endpoint.

**Alternative considered:** Encrypted JWT as state. More complex, no meaningful benefit for this flow.

### 5. No `UserRepository` interface — concrete store

Following the flat architecture philosophy, `service/auth_service.go` depends directly on `*store.UserStore` (concrete struct), not an interface. If tests need mocking, the test file defines a narrow interface locally (consumer-defined interfaces, Go proverb: "accept interfaces, return structs").

**Alternative considered:** `UserRepository` interface in `model/`. Creates a premature abstraction — there's only one implementation (PostgreSQL) and likely will only ever be one.

### 6. OAuth HTTP client with 10s timeout

Every OAuth provider adapter creates its own `http.Client{Timeout: 10 * time.Second}` instead of using `http.DefaultClient`. This prevents indefinite hangs during token exchange or profile fetch (100go.co #81).

## Risks / Trade-offs

- **[Risk] JWT secret in env var** → Acceptable for single-server deployment. Document that production must set a strong random secret.
- **[Risk] No refresh tokens** → Users re-authenticate every 24h. Acceptable for MVP; refresh can be added later without schema changes.
- **[Risk] Email-based account linking** → A user with different emails across providers gets separate accounts. This is intentional — email is the identity anchor.
- **[Trade-off] HS256 vs RS256** → Simpler setup, but means the JWT secret can't be shared with untrusted parties. Fine for a monolith.
- **[Trade-off] No interface for UserStore** → Slightly harder to unit test auth service in isolation. Mitigated by integration tests against a test database or consumer-defined mock interfaces in test files.
