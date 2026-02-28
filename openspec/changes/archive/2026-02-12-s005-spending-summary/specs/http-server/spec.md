## ADDED Requirements

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
