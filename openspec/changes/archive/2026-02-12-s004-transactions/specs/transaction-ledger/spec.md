## ADDED Requirements

### Requirement: Transaction entity with typed ID and entry splitting
The `model` package SHALL define a `Transaction` entity with typed `TransactionID`, `TransactionSource` enum (manual|import), mandatory workspace scope, and a `ValidateEntries()` method that ensures entry amounts sum to the transaction total.

#### Scenario: Transaction fields
- **WHEN** a `Transaction` struct is created
- **THEN** it SHALL have fields: `ID TransactionID`, `WorkspaceID WorkspaceID`, `CreatedBy UserID`, `TotalAmount Money`, `Description string`, `Date time.Time`, `Fingerprint *string`, `Source TransactionSource`, `Notes string`, `Entries []Entry`, `CreatedAt time.Time`, `UpdatedAt time.Time`

#### Scenario: Valid entries sum
- **WHEN** `ValidateEntries()` is called on a Transaction whose entries sum to the total amount
- **THEN** it SHALL return no error

#### Scenario: Entry sum mismatch
- **WHEN** `ValidateEntries()` is called on a Transaction whose entries do NOT sum to the total amount
- **THEN** it SHALL return `model.ErrEntrySumMismatch`

#### Scenario: No entries
- **WHEN** `ValidateEntries()` is called on a Transaction with zero entries
- **THEN** it SHALL return `model.ErrTransactionNoEntries`

### Requirement: Entry entity with category and optional participant
The `model` package SHALL define an `Entry` entity with typed `EntryID`, category assignment, optional participant, and a `Money` amount.

#### Scenario: Entry fields
- **WHEN** an `Entry` struct is created
- **THEN** it SHALL have fields: `ID EntryID`, `TransactionID TransactionID`, `CategoryID CategoryID`, `ParticipantID *UserID`, `Amount Money`, `Note string`

### Requirement: Transaction database migration
The database SHALL have `transactions` and `entries` tables with proper indexes and constraints.

#### Scenario: Transactions table schema
- **WHEN** migration 000004 is applied
- **THEN** the `transactions` table SHALL have columns: `id` (BIGSERIAL PK), `workspace_id` (BIGINT NOT NULL FK to workspaces ON DELETE CASCADE), `created_by` (BIGINT NOT NULL FK to users), `total_amount` (NUMERIC(19,4) NOT NULL), `currency` (TEXT NOT NULL DEFAULT 'PLN'), `description` (TEXT NOT NULL), `date` (DATE NOT NULL), `fingerprint` (TEXT), `source` (transaction_source NOT NULL), `notes` (TEXT NOT NULL DEFAULT ''), `created_at`, `updated_at`
- **AND** a composite index on `(workspace_id, date)` SHALL exist
- **AND** a partial unique index on `(workspace_id, fingerprint) WHERE fingerprint IS NOT NULL` SHALL exist

#### Scenario: Entries table schema
- **WHEN** migration 000004 is applied
- **THEN** the `entries` table SHALL have columns: `id` (BIGSERIAL PK), `transaction_id` (BIGINT NOT NULL FK to transactions ON DELETE CASCADE), `category_id` (BIGINT NOT NULL FK to categories), `participant_id` (BIGINT FK to users), `amount` (NUMERIC(19,4) NOT NULL), `currency` (TEXT NOT NULL DEFAULT 'PLN'), `note` (TEXT NOT NULL DEFAULT '')

#### Scenario: Migration rollback
- **WHEN** migration 000004 is rolled back
- **THEN** the `entries` and `transactions` tables and `transaction_source` enum SHALL be dropped

### Requirement: Transaction store with atomic create
The `store` package SHALL provide transaction persistence with atomic creation (transaction + entries in a single DB transaction).

#### Scenario: Atomic create
- **WHEN** a transaction with entries is created
- **THEN** both the transaction row and all entry rows SHALL be inserted within a single database transaction
- **AND** if any insert fails, all changes SHALL be rolled back

#### Scenario: Get transaction with entries
- **WHEN** `GetByID(ctx, txID)` is called for an existing transaction
- **THEN** it SHALL return the transaction with all its entries loaded

#### Scenario: Transaction not found
- **WHEN** `GetByID(ctx, nonExistentID)` is called
- **THEN** it SHALL return `model.ErrTransactionNotFound`

#### Scenario: List by workspace with date range and pagination
- **WHEN** `ListByWorkspace(ctx, wsID, from, to, limit, offset)` is called
- **THEN** it SHALL return transactions where `date >= from AND date < to`, ordered by date descending, with the given limit and offset

#### Scenario: Delete transaction
- **WHEN** `Delete(ctx, txID)` is called
- **THEN** the transaction and its entries SHALL be removed (CASCADE)

### Requirement: Transaction service with permission enforcement
The `service` package SHALL provide transaction operations with role-based permission checks.

#### Scenario: Create transaction requires editor role
- **WHEN** a viewer attempts to create a transaction
- **THEN** it SHALL return `model.ErrInsufficientPermission`

#### Scenario: Create validates entry sum
- **WHEN** `Create(ctx, userID, wsID, input)` is called with entries that don't sum to total
- **THEN** it SHALL return `model.ErrEntrySumMismatch`

#### Scenario: List transactions requires viewer role
- **WHEN** a workspace member lists transactions
- **THEN** transactions SHALL be returned regardless of role

#### Scenario: Delete transaction requires editor role
- **WHEN** a viewer attempts to delete a transaction
- **THEN** it SHALL return `model.ErrInsufficientPermission`

#### Scenario: Non-member cannot access transactions
- **WHEN** a non-member attempts any transaction operation
- **THEN** it SHALL return `model.ErrNotWorkspaceMember`
