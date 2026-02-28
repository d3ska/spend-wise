## MODIFIED Requirements

### Requirement: JSON error response helper
A `respondCodedError` function SHALL write error responses in the structured envelope format containing both a machine-readable code and a human-readable message.

#### Scenario: Coded error response format
- **WHEN** `respondCodedError(w, 400, "WORKSPACE_NAME_REQUIRED", "workspace name is required")` is called
- **THEN** the response body SHALL be `{"code":"WORKSPACE_NAME_REQUIRED","message":"workspace name is required"}`
- **AND** the status code SHALL be `400`
- **AND** `Content-Type` SHALL be `application/json`

#### Scenario: Internal server error format
- **WHEN** `respondCodedError(w, 500, "INTERNAL_ERROR", "internal server error")` is called
- **THEN** the response body SHALL be `{"code":"INTERNAL_ERROR","message":"internal server error"}`

### Requirement: Global middleware stack
The Chi router SHALL apply global middleware for all routes, including JWT authentication for protected route groups. All resource routes (workspaces, categories, transactions, summary) SHALL be registered under the protected `/api/v1/` group. The `PATCH /api/v1/auth/me` endpoint SHALL be registered within the protected route group.

#### Scenario: Public routes without auth
- **WHEN** `GET /healthz`, `GET /api/v1/auth/{provider}/url`, or `POST /api/v1/auth/{provider}/callback` is called without a JWT
- **THEN** the request SHALL be processed normally (no 401)

#### Scenario: Protected routes require JWT
- **WHEN** any route under `/api/v1/` (except auth routes) is called without a valid `Authorization: Bearer <token>` header
- **THEN** the response SHALL be `401 Unauthorized` with body `{"code":"UNAUTHORIZED","message":"missing or invalid token"}`

#### Scenario: Valid JWT injects user context
- **WHEN** a request with a valid `Authorization: Bearer <token>` header reaches a protected route
- **THEN** the `UserID` from the JWT claims SHALL be available in the request context

#### Scenario: Middleware applied
- **WHEN** any request is processed
- **THEN** the following Chi middleware SHALL be active: `RequestID`, `RealIP`, `Logger`, `Recoverer`

#### Scenario: PATCH /api/v1/auth/me route registered
- **WHEN** `PATCH /api/v1/auth/me` is called with a valid JWT
- **THEN** the request SHALL be routed to the auth handler's profile update method
