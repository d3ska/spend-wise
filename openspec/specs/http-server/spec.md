### Requirement: Chi-based HTTP server with configurable timeouts
The HTTP server SHALL use `go-chi/chi/v5` as the router and SHALL configure read/write timeouts from the application config.

#### Scenario: Server starts on configured address
- **WHEN** the server starts with `SERVER_HOST=0.0.0.0` and `SERVER_PORT=8080`
- **THEN** it SHALL listen on `0.0.0.0:8080`

#### Scenario: Timeouts are applied
- **WHEN** the server starts with `SERVER_READ_TIMEOUT=10s` and `SERVER_WRITE_TIMEOUT=30s`
- **THEN** `http.Server.ReadTimeout` SHALL be `10s` and `WriteTimeout` SHALL be `30s`

### Requirement: Health check endpoint
The server SHALL expose a `GET /healthz` endpoint that returns the server status without authentication.

#### Scenario: Health check returns OK
- **WHEN** `GET /healthz` is called
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL be `{"status":"ok"}`
- **AND** the `Content-Type` header SHALL be `application/json`

### Requirement: Graceful shutdown on signals
The server SHALL shut down gracefully when receiving `SIGINT` or `SIGTERM`, waiting up to 10 seconds for in-flight requests to complete.

#### Scenario: Graceful shutdown on SIGINT
- **WHEN** the server receives `SIGINT`
- **THEN** it SHALL stop accepting new connections
- **AND** it SHALL wait for in-flight requests to complete (up to 10 seconds)
- **AND** it SHALL exit with code 0

#### Scenario: Forced shutdown after timeout
- **WHEN** the server receives `SIGINT` and in-flight requests do not complete within 10 seconds
- **THEN** it SHALL force-close connections and exit

### Requirement: Structured logging with slog
The server SHALL use `log/slog` for all log output. Log messages SHALL include structured fields (key-value pairs).

#### Scenario: Server startup logged
- **WHEN** the server starts
- **THEN** it SHALL log `"server starting"` with the `addr` field

#### Scenario: Shutdown logged
- **WHEN** a shutdown signal is received
- **THEN** it SHALL log `"shutdown signal received"` with the `signal` field
- **AND** after shutdown completes, it SHALL log `"server stopped gracefully"`

### Requirement: JSON response helper
A `respondJSON` function SHALL write JSON responses with proper Content-Type header and status code.

#### Scenario: JSON response with data
- **WHEN** `respondJSON(w, 200, data)` is called
- **THEN** the response SHALL have `Content-Type: application/json`
- **AND** the status code SHALL be `200`
- **AND** the body SHALL be the JSON-encoded data

### Requirement: JSON error response helper
A `respondError` function SHALL write error responses in a consistent envelope format.

#### Scenario: Error response format
- **WHEN** `respondError(w, 400, "bad input")` is called
- **THEN** the response body SHALL be `{"error":"bad input"}`
- **AND** the status code SHALL be `400`

### Requirement: JSON request decoder with limits
A `decodeJSON` function SHALL decode request bodies with a 1MB size limit and reject unknown fields.

#### Scenario: Valid JSON decoded
- **WHEN** a request body contains valid JSON matching the target struct
- **THEN** `decodeJSON` SHALL populate the struct and return no error

#### Scenario: Body exceeds size limit
- **WHEN** a request body exceeds 1MB
- **THEN** `decodeJSON` SHALL return an error

#### Scenario: Unknown fields rejected
- **WHEN** a request body contains fields not in the target struct
- **THEN** `decodeJSON` SHALL return an error

### Requirement: Global middleware stack
The Chi router SHALL apply global middleware for all routes, including JWT authentication for protected route groups. All resource routes (workspaces, categories, transactions, summary) SHALL be registered under the protected `/api/v1/` group.

#### Scenario: Public routes without auth
- **WHEN** `GET /healthz`, `GET /api/v1/auth/{provider}/url`, or `POST /api/v1/auth/{provider}/callback` is called without a JWT
- **THEN** the request SHALL be processed normally (no 401)

#### Scenario: Protected routes require JWT
- **WHEN** any route under `/api/v1/` (except auth routes) is called without a valid `Authorization: Bearer <token>` header
- **THEN** the response SHALL be `401 Unauthorized` with body `{"error":"missing or invalid token"}`

#### Scenario: Valid JWT injects user context
- **WHEN** a request with a valid `Authorization: Bearer <token>` header reaches a protected route
- **THEN** the `UserID` from the JWT claims SHALL be available in the request context

#### Scenario: Middleware applied
- **WHEN** any request is processed
- **THEN** the following Chi middleware SHALL be active: `RequestID`, `RealIP`, `Logger`, `Recoverer`

### Requirement: Entry point uses run() pattern
The `main()` function SHALL delegate to a `run()` function that returns an error. `main()` SHALL log the error and call `os.Exit(1)` on failure.

#### Scenario: Startup error handling
- **WHEN** `run()` returns an error (e.g., database connection failure)
- **THEN** `main()` SHALL log the error via `slog.Error` and exit with code 1

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

### Requirement: Workspace CRUD endpoints
The server SHALL expose workspace management endpoints under `/api/v1/workspaces`.

#### Scenario: Create workspace
- **WHEN** `POST /api/v1/workspaces` is called with `{"name":"Home","description":"Household expenses","type":"shared"}`
- **THEN** the response status SHALL be `201 Created`
- **AND** the response body SHALL contain the created workspace with an `id`
- **AND** the authenticated user SHALL be added as owner

#### Scenario: List user workspaces
- **WHEN** `GET /api/v1/workspaces` is called
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL contain all workspaces where the user is a member

#### Scenario: Get workspace
- **WHEN** `GET /api/v1/workspaces/{id}` is called by a member
- **THEN** the response status SHALL be `200 OK`

#### Scenario: Update workspace
- **WHEN** `PUT /api/v1/workspaces/{id}` is called by an editor or owner
- **THEN** the response status SHALL be `200 OK`

#### Scenario: Delete workspace
- **WHEN** `DELETE /api/v1/workspaces/{id}` is called by the owner
- **THEN** the response status SHALL be `204 No Content`

### Requirement: Workspace member management endpoints
The server SHALL expose member management endpoints nested under workspaces.

#### Scenario: Add member
- **WHEN** `POST /api/v1/workspaces/{id}/members` is called by the owner with `{"user_id":5,"role":"editor"}`
- **THEN** the response status SHALL be `201 Created`

#### Scenario: Remove member
- **WHEN** `DELETE /api/v1/workspaces/{id}/members/{userID}` is called by the owner
- **THEN** the response status SHALL be `204 No Content`

### Requirement: Category CRUD endpoints
The server SHALL expose category endpoints nested under workspaces.

#### Scenario: Create category
- **WHEN** `POST /api/v1/workspaces/{id}/categories` is called by an editor with `{"name":"Groceries","icon":"cart"}`
- **THEN** the response status SHALL be `201 Created`

#### Scenario: List categories
- **WHEN** `GET /api/v1/workspaces/{id}/categories` is called by any member
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL contain all categories for the workspace

#### Scenario: Update category
- **WHEN** `PUT /api/v1/workspaces/{id}/categories/{catID}` is called by an editor
- **THEN** the response status SHALL be `200 OK`

#### Scenario: Delete category
- **WHEN** `DELETE /api/v1/workspaces/{id}/categories/{catID}` is called by an editor
- **THEN** the response status SHALL be `204 No Content`

### Requirement: Transaction CRUD endpoints
The server SHALL expose transaction management endpoints under `/api/v1/workspaces/{id}/transactions`.

#### Scenario: Create transaction
- **WHEN** `POST /api/v1/workspaces/{id}/transactions` is called by an editor with a valid transaction body including entries
- **THEN** the response status SHALL be `201 Created`
- **AND** the response body SHALL contain the created transaction with entries

#### Scenario: List transactions with date range
- **WHEN** `GET /api/v1/workspaces/{id}/transactions?from=2026-02-01&to=2026-03-01&limit=50&offset=0` is called by a member
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL contain transactions within the date range

#### Scenario: Get transaction with entries
- **WHEN** `GET /api/v1/workspaces/{id}/transactions/{txID}` is called by a member
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL include the transaction with all entries

#### Scenario: Delete transaction
- **WHEN** `DELETE /api/v1/workspaces/{id}/transactions/{txID}` is called by an editor
- **THEN** the response status SHALL be `204 No Content`

### Requirement: Transaction import endpoint
The server SHALL expose an import endpoint for batch transaction creation with deduplication.

#### Scenario: Import transactions
- **WHEN** `POST /api/v1/workspaces/{id}/transactions/import` is called by an editor with a JSON array of transactions
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL contain `imported`, `skipped` counts and the list of created transactions

### Requirement: Summary endpoint routing
The router SHALL register `GET /api/v1/workspaces/{id}/summary` under the protected route group (JWT required). The route SHALL be handled by `SummaryHandler.GetSummary`.

#### Scenario: Summary route accessible
- **WHEN** an authenticated user sends `GET /api/v1/workspaces/1/summary?from=2026-02-01&to=2026-03-01`
- **THEN** the request SHALL be routed to the summary handler

### Requirement: Summary query parameter validation
The handler SHALL parse `from` and `to` query parameters as dates in `YYYY-MM-DD` format. The handler SHALL return HTTP 400 if either parameter is missing or malformed.

#### Scenario: Missing from parameter
- **WHEN** `from` is not provided
- **THEN** the handler SHALL return HTTP 400 with an error message

#### Scenario: Invalid date format
- **WHEN** `from=not-a-date`
- **THEN** the handler SHALL return HTTP 400 with an error message

### Requirement: Summary JSON response shape
The handler SHALL return a JSON object with keys: `workspace_id`, `from`, `to`, `total_spent` (money object), `by_category` (array of category spending objects), `by_participant` (array of participant spending objects). Money objects SHALL use `{"amount": "string", "currency": "string"}` format.

#### Scenario: Full response structure
- **WHEN** the summary is successfully computed
- **THEN** the JSON response SHALL include all required fields with correct types
