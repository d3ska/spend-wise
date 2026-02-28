## ADDED Requirements

### Requirement: Bank account entity
The `model` package SHALL define a `BankAccount` entity with typed `BankAccountID`, parent connection reference, GoCardless external ID, and sync tracking.

#### Scenario: BankAccount fields
- **WHEN** a `BankAccount` struct is created
- **THEN** it SHALL have fields: `ID BankAccountID`, `BankConnectionID BankConnectionID`, `ExternalID string`, `IBAN *string`, `Name string`, `LastSyncedAt *time.Time`, `CreatedAt time.Time`, `UpdatedAt time.Time`

### Requirement: Bank account database migration
The database SHALL have `bank_accounts` and `workspace_bank_accounts` tables.

#### Scenario: Bank accounts table schema
- **WHEN** the bank accounts migration is applied
- **THEN** the `bank_accounts` table SHALL have columns: `id` (BIGSERIAL PK), `bank_connection_id` (BIGINT NOT NULL FK to bank_connections ON DELETE CASCADE), `external_id` (TEXT NOT NULL UNIQUE), `iban` (TEXT), `name` (TEXT NOT NULL DEFAULT ''), `last_synced_at` (TIMESTAMPTZ), `created_at` (TIMESTAMPTZ NOT NULL DEFAULT NOW()), `updated_at` (TIMESTAMPTZ NOT NULL DEFAULT NOW())

#### Scenario: Workspace bank accounts join table schema
- **WHEN** the bank accounts migration is applied
- **THEN** the `workspace_bank_accounts` table SHALL have columns: `workspace_id` (BIGINT NOT NULL FK to workspaces ON DELETE CASCADE), `bank_account_id` (BIGINT NOT NULL FK to bank_accounts ON DELETE CASCADE)
- **AND** a UNIQUE constraint on `(workspace_id, bank_account_id)` SHALL exist

#### Scenario: Migration rollback
- **WHEN** the bank accounts migration is rolled back
- **THEN** the `workspace_bank_accounts` and `bank_accounts` tables SHALL be dropped

### Requirement: Bank account store operations
The `store` package SHALL provide persistence for bank accounts and workspace linking.

#### Scenario: Create bank account
- **WHEN** `CreateBankAccount(ctx, model.BankAccount)` is called
- **THEN** it SHALL insert a row and return the created `BankAccount` with a generated ID

#### Scenario: Create bank account with duplicate external ID
- **WHEN** `CreateBankAccount(ctx, bankAccount)` is called with an `external_id` that already exists
- **THEN** it SHALL return `model.ErrBankAccountAlreadyExists`

#### Scenario: List bank accounts by connection
- **WHEN** `ListBankAccountsByConnection(ctx, connectionID)` is called
- **THEN** it SHALL return all bank accounts belonging to that connection

#### Scenario: Link bank account to workspace
- **WHEN** `LinkToWorkspace(ctx, workspaceID, bankAccountID)` is called
- **THEN** it SHALL insert a row in `workspace_bank_accounts`

#### Scenario: Link duplicate is idempotent
- **WHEN** `LinkToWorkspace(ctx, workspaceID, bankAccountID)` is called for an already-linked pair
- **THEN** it SHALL succeed without error (ON CONFLICT DO NOTHING)

#### Scenario: Unlink bank account from workspace
- **WHEN** `UnlinkFromWorkspace(ctx, workspaceID, bankAccountID)` is called
- **THEN** it SHALL remove the row from `workspace_bank_accounts`

#### Scenario: List bank accounts linked to workspace
- **WHEN** `ListByWorkspace(ctx, workspaceID)` is called
- **THEN** it SHALL return all bank accounts linked to that workspace, including their parent connection's `institution_name`

#### Scenario: List workspaces linked to bank account
- **WHEN** `ListWorkspacesByBankAccount(ctx, bankAccountID)` is called
- **THEN** it SHALL return all workspace IDs linked to that bank account

#### Scenario: Update last synced timestamp
- **WHEN** `UpdateLastSyncedAt(ctx, bankAccountID, timestamp)` is called
- **THEN** it SHALL update the `last_synced_at` and `updated_at` fields

### Requirement: Bank account linking service
The `service` package SHALL provide operations for linking bank accounts to workspaces with permission checks.

#### Scenario: Link bank account to workspace requires editor role
- **WHEN** a viewer attempts to link a bank account to a workspace
- **THEN** it SHALL return `model.ErrInsufficientPermission`

#### Scenario: Link requires bank account ownership
- **WHEN** a user attempts to link a bank account they do not own (via connection)
- **THEN** it SHALL return `model.ErrBankConnectionNotFound`

#### Scenario: Unlink bank account from workspace requires editor role
- **WHEN** a viewer attempts to unlink a bank account from a workspace
- **THEN** it SHALL return `model.ErrInsufficientPermission`

#### Scenario: List linked accounts requires viewer role
- **WHEN** a non-member attempts to list bank accounts for a workspace
- **THEN** it SHALL return `model.ErrNotWorkspaceMember`

### Requirement: Bank account linking HTTP handler
The `handler` package SHALL expose REST endpoints for managing bank account-workspace links.

#### Scenario: Link bank account endpoint
- **WHEN** `POST /api/v1/workspaces/{id}/bank-accounts` is called with `{"bank_account_id": 42}`
- **THEN** it SHALL link the bank account to the workspace
- **AND** it SHALL respond with `201 Created`

#### Scenario: Unlink bank account endpoint
- **WHEN** `DELETE /api/v1/workspaces/{id}/bank-accounts/{bankAccountId}` is called
- **THEN** it SHALL unlink the bank account from the workspace
- **AND** it SHALL respond with `204 No Content`

#### Scenario: List linked bank accounts endpoint
- **WHEN** `GET /api/v1/workspaces/{id}/bank-accounts` is called
- **THEN** it SHALL return all bank accounts linked to the workspace, each including `institution_name`, `iban`, `name`, and `last_synced_at`
