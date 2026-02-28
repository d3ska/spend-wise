## MODIFIED Requirements

### Requirement: Transaction entity with typed ID and entry splitting
The `model` package SHALL define a `Transaction` entity with typed `TransactionID`, `TransactionSource` enum (manual|import|bank), mandatory workspace scope, optional `BankAccountID` reference, and a `ValidateEntries()` method that ensures entry amounts sum to the transaction total.

#### Scenario: Transaction fields
- **WHEN** a `Transaction` struct is created
- **THEN** it SHALL have fields: `ID TransactionID`, `WorkspaceID WorkspaceID`, `CreatedBy UserID`, `TotalAmount Money`, `Description string`, `Date time.Time`, `Fingerprint *string`, `Source TransactionSource`, `BankAccountID *BankAccountID`, `Notes string`, `Entries []Entry`, `CreatedAt time.Time`, `UpdatedAt time.Time`

#### Scenario: Valid entries sum
- **WHEN** `ValidateEntries()` is called on a Transaction whose entries sum to the total amount
- **THEN** it SHALL return no error

#### Scenario: Entry sum mismatch
- **WHEN** `ValidateEntries()` is called on a Transaction whose entries do NOT sum to the total amount
- **THEN** it SHALL return `model.ErrEntrySumMismatch`

#### Scenario: No entries
- **WHEN** `ValidateEntries()` is called on a Transaction with zero entries
- **THEN** it SHALL return `model.ErrTransactionNoEntries`

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

#### Scenario: Bank sync migration adds bank_account_id and bank source
- **WHEN** the bank sync migration is applied
- **THEN** it SHALL add `bank_account_id` (BIGINT FK to bank_accounts ON DELETE SET NULL) to the `transactions` table
- **AND** it SHALL add `'bank'` to the `transaction_source` enum

#### Scenario: Bank sync migration rollback
- **WHEN** the bank sync migration is rolled back
- **THEN** it SHALL remove the `bank_account_id` column from `transactions`
- **AND** it SHALL remove `'bank'` from the `transaction_source` enum
