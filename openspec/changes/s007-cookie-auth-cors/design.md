## Context

SpendWise currently returns JWTs in JSON response bodies from the OAuth callback endpoint. The frontend stores these tokens in JavaScript-accessible storage and attaches them via `Authorization: Bearer` headers. This is the typical pattern for API-first backends but is not suitable for a browser-based SPA where XSS vulnerabilities could leak tokens.

The frontend will run on a separate origin (`:3000` in dev, potentially a different subdomain in production), requiring CORS configuration for cross-origin API requests.

## Goals / Non-Goals

**Goals:**
- Deliver auth tokens via httpOnly cookies that JavaScript cannot access
- Support cross-origin requests from the frontend with credentials
- Provide a logout endpoint that clears the session cookie
- Maintain backward compatibility with `Authorization: Bearer` header for non-browser clients

**Non-Goals:**
- Refresh token rotation (future enhancement — current JWT has 24h expiry)
- CSRF token middleware (SameSite=Lax on the cookie is sufficient for state-changing requests that use proper HTTP methods)
- Cookie-based auth for the dev bypass mode (dev bypass remains as-is)

## Decisions

### Decision 1: Cookie attributes

Use `Set-Cookie` with:
- `HttpOnly`: prevents JavaScript access (XSS protection)
- `Secure`: cookie only sent over HTTPS (disabled in dev via `COOKIE_SECURE=false`)
- `SameSite=Lax`: browser sends cookie on top-level navigations and same-site requests, blocks cross-site POST (CSRF protection)
- `Path=/api`: cookie only sent to API routes, not to static file routes
- `Max-Age` matches JWT expiry from config

**Alternatives considered:**
- `SameSite=Strict`: would break the OAuth redirect flow (browser navigating back from Google/GitHub is a cross-site navigation)
- `SameSite=None`: requires Secure, offers no CSRF protection — only needed for true cross-site embedding which we don't need

### Decision 2: Cookie name

Use `sw_token` as the cookie name. Short, namespaced, avoids collision.

### Decision 3: JWT middleware dual-read

The middleware will check for the JWT in this order:
1. `Cookie` header — extract `sw_token` cookie
2. `Authorization: Bearer` header — fallback

This allows browser clients to use cookies while API clients (mobile, tests, scripts) continue using Bearer tokens.

### Decision 4: CORS via go-chi/cors

Use the official `go-chi/cors` middleware with:
- `AllowedOrigins`: from `CORS_ALLOWED_ORIGINS` env var (comma-separated)
- `AllowCredentials: true` (required for cookies)
- `AllowedMethods`: GET, POST, PUT, DELETE, PATCH, OPTIONS
- `AllowedHeaders`: Content-Type, Authorization
- `MaxAge`: 300 seconds (cache preflight for 5 minutes)

CORS middleware must be applied globally (before the router groups) so preflight OPTIONS requests are handled.

### Decision 5: Logout clears cookie

`POST /api/v1/auth/logout` sets the same cookie with `Max-Age=0` (immediate expiry). This is a public route (no auth required — clearing an invalid/expired cookie is harmless).

### Decision 6: Config additions

| Env Var | Type | Default | Description |
|---------|------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | string (comma-sep) | `http://localhost:3000` | Allowed CORS origins |
| `COOKIE_DOMAIN` | string | `` (empty = current host) | Cookie domain |
| `COOKIE_SECURE` | bool | `true` | Set Secure flag on cookie |

## Risks / Trade-offs

- **[Risk] OAuth redirect with SameSite=Lax** — After OAuth provider redirects back to the frontend, the frontend makes a POST to `/auth/{provider}/callback`. Since this is a same-origin POST (frontend JS to its own API proxy), `SameSite=Lax` will include the cookie. However, the callback is where the cookie is *set*, not read, so SameSite has no bearing on this request. → No mitigation needed.

- **[Risk] Cookie not sent in dev without HTTPS** — `Secure` flag prevents cookie transmission over HTTP. → Mitigated by `COOKIE_SECURE=false` default in `.env.example` for development.

- **[Risk] CORS misconfiguration in production** — Wildcard `*` with credentials is not allowed by browsers. → Config validation: if `AllowCredentials` is true and origins contains `*`, log a warning and refuse to start.

- **[Trade-off] Dual auth read adds complexity** — Supporting both cookie and Bearer means two code paths in middleware. → Acceptable: the logic is a simple fallback chain (5 lines), and it ensures non-browser clients are not broken.
