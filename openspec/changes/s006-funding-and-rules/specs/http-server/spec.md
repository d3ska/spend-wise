## ADDED Requirements

### Requirement: Funding endpoints
The server SHALL expose funding management endpoints under `/api/v1/workspaces/{id}/fundings`.

#### Scenario: Record funding
- **WHEN** `PUT /api/v1/workspaces/{id}/fundings` is called by an editor with `{"year_month":"2026-02","amount":{"amount":"3000.00","currency":"PLN"}}`
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL contain the upserted funding record

#### Scenario: List fundings
- **WHEN** `GET /api/v1/workspaces/{id}/fundings?from=2026-01&to=2026-03` is called by a member
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL contain funding records in the month range

### Requirement: Rule CRUD endpoints
The server SHALL expose categorization rule management endpoints under `/api/v1/workspaces/{id}/rules`.

#### Scenario: Create rule
- **WHEN** `POST /api/v1/workspaces/{id}/rules` is called by an editor with `{"match_pattern":"biedronka","target_category_id":5,"priority":10}`
- **THEN** the response status SHALL be `201 Created`

#### Scenario: Update rule
- **WHEN** `PUT /api/v1/workspaces/{id}/rules/{ruleID}` is called by an editor
- **THEN** the response status SHALL be `200 OK`

#### Scenario: Delete rule
- **WHEN** `DELETE /api/v1/workspaces/{id}/rules/{ruleID}` is called by an editor
- **THEN** the response status SHALL be `204 No Content`

### Requirement: Rule toggle endpoint
The server SHALL expose a toggle endpoint for enabling/disabling rules.

#### Scenario: Toggle rule enabled
- **WHEN** `PATCH /api/v1/workspaces/{id}/rules/{ruleID}/toggle` is called by an editor with `{"enabled": false}`
- **THEN** the response status SHALL be `200 OK`
- **AND** the rule's enabled status SHALL be updated
