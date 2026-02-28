## MODIFIED Requirements

### Requirement: Transaction entity with typed ID and entry splitting
The `model` package SHALL define a `Transaction` entity with typed `TransactionID`, `TransactionSource` enum (manual|import|bank), mandatory workspace scope, and a `ValidateEntries()` method that ensures entry amounts sum to the transaction total.

#### Scenario: Transaction fields
- **WHEN** a `Transaction` struct is created
- **THEN** it SHALL have fields: `ID TransactionID`, `WorkspaceID WorkspaceID`, `CreatedBy UserID`, `TotalAmount Money`, `Description string`, `Date time.Time`, `Fingerprint *string`, `Source TransactionSource`, `Type TransactionType`, `Notes string`, `Entries []Entry`, `BankAccountID *BankAccountID`, `IBAN *string`, `CreatedAt time.Time`, `UpdatedAt time.Time`

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
- **WHEN** migration 000011 is applied
- **THEN** the `transactions` table SHALL have an additional column: `iban` (TEXT, nullable)
- **AND** existing bank-synced transactions SHALL have `iban` backfilled from their linked `bank_accounts` row

#### Scenario: Migration rollback
- **WHEN** migration 000011 is rolled back
- **THEN** the `iban` column SHALL be dropped from `transactions`
- **AND** the `idx_bank_accounts_iban_currency` index SHALL be dropped

## ADDED Requirements

### Requirement: IBAN exposed in transaction API responses
The transaction API responses SHALL include the `iban` field when present.

#### Scenario: Transaction response includes IBAN
- **WHEN** a transaction with a non-NULL `iban` is returned from the API
- **THEN** the response JSON SHALL include `"iban": "<value>"`

#### Scenario: Transaction response without IBAN
- **WHEN** a transaction with a NULL `iban` is returned from the API
- **THEN** the response JSON SHALL either omit the `iban` field or include `"iban": null`
