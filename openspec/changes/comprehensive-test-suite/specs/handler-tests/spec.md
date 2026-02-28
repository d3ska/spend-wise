## ADDED Requirements

### Requirement: Handler test infrastructure
Handler tests SHALL use `httptest.NewRequest` and `httptest.NewRecorder`. Auth context SHALL be injected via a test helper that sets the user ID in the request context. Chi URL params SHALL be injected via `chi.NewRouteContext`. Each test SHALL create a fresh recorder.

#### Scenario: Authenticated request setup
- **WHEN** a handler test needs an authenticated request
- **THEN** the test helper injects the user ID into `r.Context()` using `auth.withUserID` (exposed via test export)

#### Scenario: URL parameter injection
- **WHEN** a handler test needs a chi URL param (e.g., `{id}`)
- **THEN** the test sets it via `chi.NewRouteContext()` attached to the request context

### Requirement: Request decoding tests
Tests SHALL verify `decodeJSON` behavior: valid JSON, invalid JSON, unknown fields rejected, body exceeding 1MB limit.

#### Scenario: Valid JSON body
- **WHEN** `decodeJSON` is called with a valid JSON body matching the target struct
- **THEN** the struct is populated and nil error is returned

#### Scenario: Invalid JSON body
- **WHEN** `decodeJSON` is called with malformed JSON
- **THEN** a non-nil error is returned

#### Scenario: Unknown fields rejected
- **WHEN** `decodeJSON` is called with JSON containing fields not in the target struct
- **THEN** a non-nil error is returned (because `DisallowUnknownFields` is set)

#### Scenario: Body size limit
- **WHEN** `decodeJSON` is called with a body exceeding 1MB
- **THEN** a non-nil error is returned

### Requirement: Response helper tests
Tests SHALL verify `respondJSON` sets Content-Type header, writes the correct status code, and encodes the body as JSON. Tests SHALL verify `respondError` produces the `{"error": "message"}` envelope.

#### Scenario: respondJSON sets headers and body
- **WHEN** `respondJSON(w, 200, data)` is called
- **THEN** `w` has Content-Type `application/json`, status 200, and body is JSON-encoded `data`

#### Scenario: respondError envelope format
- **WHEN** `respondError(w, 400, "bad input")` is called
- **THEN** `w` has status 400 and body is `{"error":"bad input"}`

### Requirement: handleServiceError mapping tests
Tests SHALL verify that `handleServiceError` maps each sentinel error to the correct HTTP status code. Table-driven test with all error→status mappings.

#### Scenario: Not found errors map to 404
- **WHEN** `handleServiceError` is called with `ErrWorkspaceNotFound`, `ErrCategoryNotFound`, `ErrFundingNotFound`, `ErrRuleNotFound`, `ErrBankConnectionNotFound`, or `ErrBankAccountNotFound`
- **THEN** the response status is 404

#### Scenario: Permission errors map to 403
- **WHEN** `handleServiceError` is called with `ErrNotWorkspaceMember` or `ErrInsufficientPermission`
- **THEN** the response status is 403

#### Scenario: Validation errors map to 400
- **WHEN** `handleServiceError` is called with `ErrWorkspaceNameRequired`, `ErrWorkspaceDescriptionRequired`, `ErrInvalidWorkspaceType`, `ErrCategoryBudgetsExceed`, or `ErrCategoryUndeletable`
- **THEN** the response status is 400

#### Scenario: Unknown errors map to 500
- **WHEN** `handleServiceError` is called with an unrecognized error
- **THEN** the response status is 500 and body contains "internal server error"

### Requirement: WorkspaceHandler endpoint tests
Tests SHALL verify the workspace handler's HTTP behavior for each endpoint: correct status codes, response bodies, error cases (unauthorized, bad ID, service errors).

#### Scenario: CreateWorkspace success
- **WHEN** POST `/workspaces` is called with valid JSON body and authenticated context
- **THEN** status is 201 and response body contains the workspace fields

#### Scenario: CreateWorkspace unauthorized
- **WHEN** POST `/workspaces` is called without auth context
- **THEN** status is 401

#### Scenario: CreateWorkspace invalid body
- **WHEN** POST `/workspaces` is called with malformed JSON
- **THEN** status is 400

#### Scenario: GetWorkspace success
- **WHEN** GET `/workspaces/{id}` is called with a valid ID and authenticated viewer
- **THEN** status is 200 and response body contains workspace data

#### Scenario: GetWorkspace not found
- **WHEN** GET `/workspaces/{id}` is called and the service returns `ErrWorkspaceNotFound`
- **THEN** status is 404

#### Scenario: DeleteWorkspace success
- **WHEN** DELETE `/workspaces/{id}` is called by an owner
- **THEN** status is 204 with no body

#### Scenario: ListWorkspaces returns array
- **WHEN** GET `/workspaces` is called by an authenticated user
- **THEN** status is 200 and response body is a JSON array

### Requirement: Other handler endpoint tests
Tests SHALL cover all remaining handlers (transaction, category, funding, rule, summary, bank, auth) following the same patterns: success cases, unauthorized, not found, validation errors.

#### Scenario: Transaction handler CRUD
- **WHEN** transaction handler endpoints are called with valid/invalid inputs
- **THEN** correct status codes and response formats are returned

#### Scenario: Auth handler callback
- **WHEN** the auth callback handler receives a valid code and state
- **THEN** it sets an auth cookie and redirects or returns user data

#### Scenario: Auth handler logout
- **WHEN** the logout handler is called
- **THEN** the auth cookie is cleared
