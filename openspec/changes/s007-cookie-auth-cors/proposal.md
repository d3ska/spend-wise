## Why

The backend currently returns JWT tokens in JSON response bodies, requiring the frontend to store them in JavaScript-accessible storage (localStorage/sessionStorage). This is vulnerable to XSS attacks. For a commercial-grade household finance app, authentication tokens must be stored in httpOnly cookies that JavaScript cannot access. Additionally, the frontend (running on a separate port/domain) needs CORS configuration to communicate with the backend API.

## What Changes

- **BREAKING**: `POST /api/v1/auth/{provider}/callback` will set an httpOnly cookie instead of returning a `token` field in the JSON body. The response body will only contain the `user` object.
- JWT middleware will read the auth token from the `Cookie` header, with fallback to `Authorization: Bearer` header for backward compatibility (e.g., mobile clients, API testing).
- Add CORS middleware (go-chi/cors) allowing configurable frontend origins with `credentials: true`.
- Add `POST /api/v1/auth/logout` endpoint that clears the auth cookie.
- Add configuration fields: `CORS_ALLOWED_ORIGINS`, `COOKIE_DOMAIN`, `COOKIE_SECURE`.

## Capabilities

### New Capabilities
- `cookie-session`: httpOnly cookie-based session management, including cookie setting on login, reading in middleware, and clearing on logout.
- `cors-config`: CORS middleware configuration for cross-origin frontend requests with credentials.

### Modified Capabilities
- `oauth-auth`: Callback endpoint response changes from `{token, user}` to `{user}` with token delivered via Set-Cookie header. JWT middleware reads from Cookie header in addition to Authorization header.
- `http-server`: New logout route added to public auth routes. CORS middleware added to global middleware stack.

## Impact

- **Code**: `auth/middleware.go`, `handler/handler_auth.go`, `handler/router.go`, `config/config.go`, `cmd/api/main.go`
- **API**: Breaking change to callback response shape (no `token` field). New `POST /api/v1/auth/logout` endpoint.
- **Dependencies**: `github.com/go-chi/cors` (new dependency)
- **Clients**: Frontend must use `credentials: "include"` on all fetch/axios requests. Existing API consumers using `Authorization: Bearer` header will continue to work.
