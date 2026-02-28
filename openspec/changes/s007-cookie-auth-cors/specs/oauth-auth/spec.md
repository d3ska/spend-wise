## MODIFIED Requirements

### Requirement: Auth endpoint - OAuth callback
The server SHALL expose `POST /api/v1/auth/{provider}/callback` that exchanges an OAuth code for a user profile and sets an httpOnly session cookie.

#### Scenario: Successful callback
- **WHEN** `POST /api/v1/auth/google/callback` is called with body `{"code":"valid-code","state":"matching-state"}`
- **THEN** the response status SHALL be `200 OK`
- **AND** the response SHALL set an httpOnly `sw_token` cookie containing the JWT
- **AND** the response body SHALL contain `{"user":{...}}` without a `token` field

#### Scenario: Missing code
- **WHEN** `POST /api/v1/auth/google/callback` is called with body `{}`
- **THEN** the response status SHALL be `400 Bad Request`

## MODIFIED Requirements

### Requirement: Global middleware stack
The Chi router SHALL apply global middleware for all routes, including CORS middleware, and JWT authentication for protected route groups. The logout route SHALL be registered as a public auth route. All resource routes (workspaces, categories, transactions, summary, fundings, rules) SHALL be registered under the protected `/api/v1/` group.

#### Scenario: Public routes without auth
- **WHEN** `GET /healthz`, `GET /api/v1/auth/{provider}/url`, `POST /api/v1/auth/{provider}/callback`, or `POST /api/v1/auth/logout` is called without a JWT
- **THEN** the request SHALL be processed normally (no 401)

#### Scenario: Protected routes require JWT
- **WHEN** any route under `/api/v1/` (except auth routes) is called without a valid JWT (neither cookie nor Bearer header)
- **THEN** the response SHALL be `401 Unauthorized` with body `{"error":"missing or invalid token"}`

#### Scenario: Valid JWT injects user context
- **WHEN** a request with a valid JWT (via cookie or Bearer header) reaches a protected route
- **THEN** the `UserID` from the JWT claims SHALL be available in the request context

#### Scenario: Middleware applied
- **WHEN** any request is processed
- **THEN** the following middleware SHALL be active: `CORS`, `RequestID`, `RealIP`, `Logger`, `Recoverer`
