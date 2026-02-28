## MODIFIED Requirements

### Requirement: Global middleware stack
The Chi router SHALL apply global middleware for all routes, including JWT authentication for protected route groups.

#### Scenario: Public routes without auth
- **WHEN** `GET /healthz`, `GET /api/v1/auth/{provider}/url`, or `POST /api/v1/auth/{provider}/callback` is called without a JWT
- **THEN** the request SHALL be processed normally (no 401)

#### Scenario: Protected routes require JWT
- **WHEN** any route under `/api/v1/` (except auth routes) is called without a valid `Authorization: Bearer <token>` header
- **THEN** the response SHALL be `401 Unauthorized` with body `{"error":"missing or invalid token"}`

#### Scenario: Valid JWT injects user context
- **WHEN** a request with a valid `Authorization: Bearer <token>` header reaches a protected route
- **THEN** the `UserID` from the JWT claims SHALL be available in the request context

## ADDED Requirements

### Requirement: Auth endpoint - get OAuth URL
The server SHALL expose `GET /api/v1/auth/{provider}/url` that returns an OAuth2 authorization URL.

#### Scenario: Get Google auth URL
- **WHEN** `GET /api/v1/auth/google/url` is called
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL contain `{"url":"https://accounts.google.com/...","state":"..."}`
- **AND** `Content-Type` SHALL be `application/json`

#### Scenario: Unsupported provider
- **WHEN** `GET /api/v1/auth/facebook/url` is called
- **THEN** the response status SHALL be `400 Bad Request`
- **AND** the response body SHALL be `{"error":"unsupported provider: facebook"}`

### Requirement: Auth endpoint - OAuth callback
The server SHALL expose `POST /api/v1/auth/{provider}/callback` that exchanges an OAuth code for a JWT and user profile.

#### Scenario: Successful callback
- **WHEN** `POST /api/v1/auth/google/callback` is called with body `{"code":"valid-code","state":"matching-state"}`
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL contain `{"token":"<jwt>","user":{...}}`

#### Scenario: Missing code
- **WHEN** `POST /api/v1/auth/google/callback` is called with body `{}`
- **THEN** the response status SHALL be `400 Bad Request`

### Requirement: Auth endpoint - current user
The server SHALL expose `GET /api/v1/auth/me` (protected) that returns the currently authenticated user.

#### Scenario: Authenticated user
- **WHEN** `GET /api/v1/auth/me` is called with a valid JWT
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL contain the user's id, email, display_name, and avatar_url

#### Scenario: Unauthenticated request
- **WHEN** `GET /api/v1/auth/me` is called without a JWT
- **THEN** the response status SHALL be `401 Unauthorized`
