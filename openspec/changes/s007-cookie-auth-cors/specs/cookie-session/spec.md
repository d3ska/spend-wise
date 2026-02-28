## ADDED Requirements

### Requirement: Set httpOnly auth cookie on successful OAuth callback
The `POST /api/v1/auth/{provider}/callback` endpoint SHALL set an httpOnly cookie named `sw_token` containing the JWT on successful authentication. The cookie SHALL have attributes: `HttpOnly=true`, `SameSite=Lax`, `Path=/api`, `Secure` controlled by `COOKIE_SECURE` config, `Domain` controlled by `COOKIE_DOMAIN` config, and `Max-Age` matching the JWT expiry duration in seconds.

#### Scenario: Successful callback sets cookie
- **WHEN** `POST /api/v1/auth/google/callback` is called with a valid `{code, state}` body
- **THEN** the response SHALL include a `Set-Cookie` header with name `sw_token`
- **AND** the cookie SHALL have `HttpOnly` flag set
- **AND** the cookie SHALL have `SameSite=Lax`
- **AND** the cookie SHALL have `Path=/api`

#### Scenario: Cookie Secure flag in production
- **WHEN** `COOKIE_SECURE=true` and a successful callback occurs
- **THEN** the `Set-Cookie` header SHALL include the `Secure` flag

#### Scenario: Cookie Secure flag in development
- **WHEN** `COOKIE_SECURE=false` and a successful callback occurs
- **THEN** the `Set-Cookie` header SHALL NOT include the `Secure` flag

#### Scenario: Response body contains only user
- **WHEN** a successful callback occurs
- **THEN** the response body SHALL contain the `user` object
- **AND** the response body SHALL NOT contain a `token` field

### Requirement: JWT middleware reads token from cookie
The JWT authentication middleware SHALL first attempt to read the token from the `sw_token` cookie. If no cookie is present, it SHALL fall back to reading from the `Authorization: Bearer` header.

#### Scenario: Authenticated via cookie
- **WHEN** a request includes a `sw_token` cookie with a valid JWT
- **AND** no `Authorization` header is present
- **THEN** the request SHALL be authenticated and `UserID` SHALL be available in context

#### Scenario: Authenticated via Bearer header (fallback)
- **WHEN** a request includes an `Authorization: Bearer <valid-jwt>` header
- **AND** no `sw_token` cookie is present
- **THEN** the request SHALL be authenticated and `UserID` SHALL be available in context

#### Scenario: Cookie takes precedence over Bearer
- **WHEN** a request includes both a `sw_token` cookie and an `Authorization: Bearer` header
- **THEN** the middleware SHALL use the cookie token

#### Scenario: No token provided
- **WHEN** a request has neither a `sw_token` cookie nor an `Authorization: Bearer` header
- **THEN** the middleware SHALL return `401 Unauthorized` with body `{"error":"missing or invalid token"}`

### Requirement: Logout endpoint clears auth cookie
The server SHALL expose `POST /api/v1/auth/logout` that clears the `sw_token` cookie by setting `Max-Age=0` with the same `Path`, `Domain`, `HttpOnly`, `SameSite`, and `Secure` attributes as the login cookie.

#### Scenario: Successful logout
- **WHEN** `POST /api/v1/auth/logout` is called
- **THEN** the response SHALL include a `Set-Cookie` header for `sw_token` with `Max-Age=0`
- **AND** the response status SHALL be `200 OK`
- **AND** the response body SHALL be `{"status":"logged_out"}`

#### Scenario: Logout without existing cookie
- **WHEN** `POST /api/v1/auth/logout` is called without an existing `sw_token` cookie
- **THEN** the response SHALL still return `200 OK` (clearing a non-existent cookie is harmless)

### Requirement: Cookie configuration from environment
The application config SHALL include fields for cookie configuration loaded from environment variables.

#### Scenario: Default cookie config
- **WHEN** no cookie-related env vars are set
- **THEN** `COOKIE_DOMAIN` SHALL default to empty string (current host)
- **AND** `COOKIE_SECURE` SHALL default to `true`

#### Scenario: Custom cookie config
- **WHEN** `COOKIE_DOMAIN=example.com` and `COOKIE_SECURE=false` are set
- **THEN** the cookie domain SHALL be `example.com`
- **AND** the Secure flag SHALL be `false`
