## MODIFIED Requirements

### Requirement: Global middleware stack
The Chi router SHALL apply global middleware for all routes, including JWT authentication for protected route groups. Transaction routes SHALL be added under the protected `/api/v1/workspaces/{id}/` group.

#### Scenario: Public routes without auth
- **WHEN** `GET /healthz`, `GET /api/v1/auth/{provider}/url`, or `POST /api/v1/auth/{provider}/callback` is called without a JWT
- **THEN** the request SHALL be processed normally (no 401)

#### Scenario: Protected routes require JWT
- **WHEN** any route under `/api/v1/` (except auth routes) is called without a valid `Authorization: Bearer <token>` header
- **THEN** the response SHALL be `401 Unauthorized` with body `{"error":"missing or invalid token"}`

## ADDED Requirements

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
