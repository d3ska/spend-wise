## MODIFIED Requirements

### Requirement: Global middleware stack
The Chi router SHALL apply global middleware for all routes, including JWT authentication for protected route groups. Workspace and category routes SHALL be added under the protected `/api/v1/` group.

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
