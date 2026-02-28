## 1. Configuration

- [x] 1.1 Add `CORSAllowedOrigins` (string, comma-separated, default `http://localhost:3000`), `CookieDomain` (string, default empty), `CookieSecure` (bool, default `true`) fields to `config.Config` struct with appropriate env tags
- [x] 1.2 Update `.env.example` with `CORS_ALLOWED_ORIGINS`, `COOKIE_DOMAIN`, `COOKIE_SECURE` entries and comments

## 2. Dependencies

- [x] 2.1 Add `github.com/go-chi/cors` dependency via `go get` and run `go mod tidy`

## 3. Cookie Auth — Middleware

- [x] 3.1 Update `auth/middleware.go` `JWTMiddleware` to read token from `sw_token` cookie first, then fall back to `Authorization: Bearer` header
- [x] 3.2 Add unit tests for dual-read middleware: cookie-only, bearer-only, both present (cookie wins), neither present (401)

## 4. Cookie Auth — Handler

- [x] 4.1 Create a `CookieConfig` struct (or add fields to `AuthHandler`) to hold `Domain`, `Secure`, `MaxAge` values needed for cookie setting
- [x] 4.2 Update `HandleCallback` in `handler/handler_auth.go` to set `sw_token` httpOnly cookie (HttpOnly, SameSite=Lax, Path=/api, Secure from config, Max-Age from JWT expiry) and return only `{user}` in response body (remove `token` field)
- [x] 4.3 Add `HandleLogout` method to `AuthHandler` that clears the `sw_token` cookie by setting Max-Age=0 with same attributes, returns `{"status":"logged_out"}`

## 5. CORS Middleware

- [x] 5.1 Add CORS middleware setup in `handler/router.go` using `go-chi/cors` with options from config: `AllowedOrigins`, `AllowCredentials: true`, `AllowedMethods` (GET, POST, PUT, DELETE, PATCH, OPTIONS), `AllowedHeaders` (Content-Type, Authorization), `MaxAge: 300`
- [x] 5.2 Ensure CORS middleware is applied globally (first in the middleware chain, before route matching)

## 6. Route Registration

- [x] 6.1 Add `POST /api/v1/auth/logout` as a public route (alongside existing auth routes, outside the JWT-protected group) in `handler/router.go`
- [x] 6.2 Pass cookie config and CORS origins through `RouterConfig` struct

## 7. Composition Root

- [x] 7.1 Update `cmd/api/main.go` to parse new config fields and pass them to `RouterConfig` and `AuthHandler`

## 8. Integration Verification

- [x] 8.1 Run full test suite (`make test`) to ensure no regressions
- [ ] 8.2 Manual smoke test: verify cookie is set on callback, protected routes accept cookie, logout clears cookie
