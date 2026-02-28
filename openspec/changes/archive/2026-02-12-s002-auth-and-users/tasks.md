## 1. User Entity

- [x] 1.1 Create `model/user.go` — `UserID` typed wrapper (int64), `AuthProvider` type with `ProviderGoogle`/`ProviderGitHub` constants, `User` struct with fields: ID, Email, DisplayName, AvatarURL, Provider, ProviderID, CreatedAt, UpdatedAt

## 2. Database Migration

- [x] 2.1 Create `migrations/000001_create_users.up.sql` — `auth_provider` enum type, `users` table with BIGSERIAL PK, email unique, display_name, avatar_url, provider, provider_id, created_at, updated_at
- [x] 2.2 Create `migrations/000001_create_users.down.sql` — drop `users` table, drop `auth_provider` enum type

## 3. User Store

- [x] 3.1 Create `db/queries/users.sql` — sqlc queries: GetUserByID, GetUserByEmail, InsertUser, UpdateUserProvider
- [x] 3.2 Run `sqlc generate` to generate Go code in `store/`
- [x] 3.3 Create `store/user_store.go` — `UserStore` struct wrapping `*Queries`, methods: `GetByID(ctx, UserID)`, `GetOrCreateByEmail(ctx, User)` (upsert logic using GetUserByEmail + InsertUser or UpdateUserProvider)

## 4. JWT Helpers

- [x] 4.1 Create `auth/jwt.go` — `SignToken(userID, secret, expiry)` returns signed JWT string, `VerifyToken(tokenStr, secret)` returns `UserID` or error. Claims: sub (user ID), exp, iat
- [x] 4.2 Create `auth/jwt_test.go` — table-driven tests: sign+verify round-trip, expired token rejected, wrong secret rejected, malformed token rejected

## 5. OAuth Providers

- [x] 5.1 Create `auth/provider.go` — `UserInfo` struct (Email, DisplayName, AvatarURL, ProviderID), `Provider` interface (AuthCodeURL, Exchange, FetchUser)
- [x] 5.2 Create `auth/google.go` — `GoogleProvider` struct implementing `Provider`, uses `golang.org/x/oauth2/google`, fetches profile from `https://www.googleapis.com/oauth2/v2/userinfo`, 10s HTTP client timeout, defer resp.Body.Close()
- [x] 5.3 Create `auth/github.go` — `GitHubProvider` struct implementing `Provider`, uses GitHub OAuth2 endpoints, fetches profile from `https://api.github.com/user`, fallback to `/user/emails` for primary email, 10s HTTP client timeout, defer resp.Body.Close()

## 6. Auth Service

- [x] 6.1 Create `service/auth_service.go` — `AuthService` struct with `providers map[model.AuthProvider]auth.Provider`, `userStore *store.UserStore`, `jwtSecret string`, `jwtExpiry time.Duration`. Method: `Authenticate(ctx, providerName, code) (token string, user model.User, err error)` — validates provider exists, calls Exchange+FetchUser, calls userStore.GetOrCreateByEmail, signs JWT

## 7. Auth Middleware

- [x] 7.1 Create `auth/middleware.go` — `JWTMiddleware(secret string)` returns Chi middleware that extracts `Authorization: Bearer <token>`, verifies JWT, puts `UserID` in context. Returns 401 on missing/invalid token
- [x] 7.2 Create `auth/context.go` — `UserIDFromContext(ctx)` helper to retrieve `UserID` from request context, `contextKey` type for type-safe context keys

## 8. Auth Handler

- [x] 8.1 Create `handler/handler_auth.go` — `AuthHandler` struct with `authService *service.AuthService`, `providers map[model.AuthProvider]auth.Provider`. Handler methods:
  - `GetAuthURL(w, r)` — reads `{provider}` from URL, generates random state, calls provider.AuthCodeURL(state), responds with `{url, state}`
  - `HandleCallback(w, r)` — reads `{provider}` from URL, decodes `{code, state}` from body, calls authService.Authenticate, responds with `{token, user}`
  - `GetMe(w, r)` — reads UserID from context, fetches user from store, responds with user profile

## 9. Router Wiring

- [x] 9.1 Update `handler/router.go` — add auth route group (`/api/v1/auth/{provider}/url`, `/api/v1/auth/{provider}/callback`, `/api/v1/auth/me`), apply JWT middleware to protected `/api/v1/` routes
- [x] 9.2 Update `cmd/api/main.go` — create OAuth providers from config, create UserStore, create AuthService, create AuthHandler, pass to NewRouter

## 10. Verification

- [x] 10.1 Run `go build ./...` — must compile with zero errors
- [x] 10.2 Run `go vet ./...` — must pass with zero warnings
- [x] 10.3 Run `go test -count=1 ./...` — all tests must pass
