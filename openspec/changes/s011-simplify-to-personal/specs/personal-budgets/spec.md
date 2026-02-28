## ADDED Requirements

### Requirement: Per-category budget storage
The system SHALL store budget targets per category per month by extending the `fundings` table with a nullable `category_id` column. A row with `category_id = NULL` represents the overall monthly budget. A row with a non-null `category_id` represents a budget target for that specific category in that month.

The unique constraint SHALL be `(workspace_id, user_id, year_month, COALESCE(category_id, 0))` to prevent duplicate entries while correctly handling NULLs.

#### Scenario: Overall budget stored
- **WHEN** a funding record is upserted with `category_id = NULL`, `year_month = "2026-02"`, and `amount = 5000.00`
- **THEN** the record is stored as the overall budget for that month
- **AND** only one overall budget per user per workspace per month SHALL exist

#### Scenario: Category budget stored
- **WHEN** a funding record is upserted with `category_id = 3` (Groceries), `year_month = "2026-02"`, and `amount = 600.00`
- **THEN** the record is stored as the budget for category 3 in that month
- **AND** it SHALL coexist with the overall budget for the same month

#### Scenario: Category budget updated
- **WHEN** a funding record is upserted with the same `(workspace_id, user_id, year_month, category_id)` as an existing record
- **THEN** the existing record's amount SHALL be updated (not duplicated)

### Requirement: Budget API uses caller identity
The funding/budget record endpoint SHALL default `user_id` to the authenticated caller's ID from the JWT token. The `user_id` field in the request body SHALL be optional; if omitted, the caller's ID is used.

#### Scenario: Budget recorded without user_id
- **WHEN** a PUT request to `/api/v1/workspaces/{id}/fundings` is sent with body `{ "year_month": "2026-02", "amount": { "amount": "5000.00", "currency": "PLN" } }` (no `user_id`, no `category_id`)
- **THEN** the system SHALL use the caller's user ID and store an overall budget

#### Scenario: Category budget recorded with category_id
- **WHEN** a PUT request to `/api/v1/workspaces/{id}/fundings` is sent with body `{ "year_month": "2026-02", "category_id": 3, "amount": { "amount": "600.00", "currency": "PLN" } }`
- **THEN** the system SHALL store a budget for category 3 for that month

#### Scenario: Legacy request with explicit user_id still works
- **WHEN** a PUT request includes `"user_id": 42` in the body
- **THEN** the system SHALL accept it (backwards-compatible) and use the provided user_id

### Requirement: Budget listing includes category information
The funding list endpoint SHALL return `category_id` (nullable) in each funding response object. When `category_id` is present, the response SHALL also include `category_name` for display convenience.

#### Scenario: List fundings returns category budgets
- **WHEN** a GET request to `/api/v1/workspaces/{id}/fundings?from=2026-01&to=2026-03` is made
- **THEN** the response SHALL include both overall budgets (`category_id: null`) and per-category budgets (`category_id: N, category_name: "..."`)

### Requirement: Delete a budget entry
The system SHALL support deleting a specific budget entry by ID via `DELETE /api/v1/workspaces/{id}/fundings/{fundingId}`. This allows users to remove a per-category budget or overall budget they no longer want.

#### Scenario: Delete category budget
- **WHEN** a DELETE request is sent to `/api/v1/workspaces/{id}/fundings/17`
- **THEN** the funding record with ID 17 SHALL be deleted
- **AND** the response status SHALL be 204 No Content

#### Scenario: Delete non-existent budget
- **WHEN** a DELETE request is sent for a funding ID that does not exist
- **THEN** the response status SHALL be 404

### Requirement: Budgets page replaces Fundings page
The frontend SHALL provide a "Budgets" page (accessible via sidebar navigation labeled "Budgets") that replaces the existing "Fundings" page. The page SHALL display:
1. A month picker defaulting to the current month
2. An overall budget input field
3. A list of all workspace categories, each with a budget input field
4. A save button that upserts all changed values for the selected month

#### Scenario: Setting budgets for current month
- **WHEN** user navigates to the Budgets page
- **THEN** the page SHALL show the current month selected, with existing budget values pre-filled (if any) and empty fields for categories without budgets

#### Scenario: Saving budget changes
- **WHEN** user modifies the overall budget and two category budgets, then clicks Save
- **THEN** the system SHALL upsert three funding records (one overall, two per-category) for the selected month
- **AND** a success toast SHALL be displayed

#### Scenario: Changing month
- **WHEN** user selects a different month from the month picker
- **THEN** the page SHALL reload budget values for that month (existing values or empty)
