## ADDED Requirements

### Requirement: Funding entity with monthly granularity
The `model` package SHALL define a `Funding` entity with typed `FundingID`, workspace scope, user scope, `YearMonth` string field (format "YYYY-MM"), and `Money` amount.

#### Scenario: Funding fields
- **WHEN** a `Funding` struct is created
- **THEN** it SHALL have fields: `ID FundingID`, `WorkspaceID WorkspaceID`, `UserID UserID`, `YearMonth string`, `Amount Money`, `CreatedAt time.Time`, `UpdatedAt time.Time`

### Requirement: Funding database migration
The database SHALL have a `fundings` table with a unique constraint on (workspace_id, user_id, year_month).

#### Scenario: Fundings table schema
- **WHEN** migration 000005 is applied
- **THEN** the `fundings` table SHALL have columns: `id` (BIGSERIAL PK), `workspace_id` (BIGINT NOT NULL FK to workspaces ON DELETE CASCADE), `user_id` (BIGINT NOT NULL FK to users), `year_month` (TEXT NOT NULL), `amount` (NUMERIC(19,4) NOT NULL), `currency` (TEXT NOT NULL DEFAULT 'PLN'), `created_at`, `updated_at`
- **AND** a unique constraint on `(workspace_id, user_id, year_month)` SHALL exist

#### Scenario: Migration rollback
- **WHEN** migration 000005 is rolled back
- **THEN** the `fundings` table SHALL be dropped

### Requirement: Funding upsert semantics
The funding store SHALL support upsert: inserting a new funding record or updating the amount if a record for the same (workspace_id, user_id, year_month) already exists.

#### Scenario: First funding for a month
- **WHEN** a funding record is created for a user/month that has no existing record
- **THEN** a new row SHALL be inserted and the created funding SHALL be returned

#### Scenario: Update existing funding
- **WHEN** a funding record is created for a user/month that already has a record
- **THEN** the existing record's amount SHALL be updated and the updated funding SHALL be returned

### Requirement: Funding list by workspace and date range
The funding store SHALL support listing fundings for a workspace filtered by a range of year_month values.

#### Scenario: List fundings in range
- **WHEN** `ListByWorkspace(ctx, wsID, fromMonth, toMonth)` is called
- **THEN** it SHALL return all fundings where `year_month >= fromMonth AND year_month <= toMonth`

### Requirement: Funding service with permission enforcement
The funding service SHALL enforce editor+ permission for recording and viewer+ for listing.

#### Scenario: Record funding requires editor role
- **WHEN** a viewer attempts to record funding
- **THEN** it SHALL return `model.ErrInsufficientPermission`

#### Scenario: List fundings requires viewer role
- **WHEN** a workspace member lists fundings
- **THEN** fundings SHALL be returned regardless of role (viewer, editor, owner)

#### Scenario: Non-member cannot access fundings
- **WHEN** a non-member attempts any funding operation
- **THEN** it SHALL return `model.ErrNotWorkspaceMember`
