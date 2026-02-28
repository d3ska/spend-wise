## MODIFIED Requirements

### Requirement: Transaction entity with typed ID and entry splitting
The `model` package SHALL define a `Transaction` entity with typed `TransactionID`, `TransactionSource` enum (manual|import|bank), `TransactionType` enum (expense|income|transfer), mandatory workspace scope, and a `ValidateEntries()` method that ensures entry amounts sum to the transaction total.

#### Scenario: Transaction fields
- **WHEN** a `Transaction` struct is created
- **THEN** it SHALL have fields: `ID TransactionID`, `WorkspaceID WorkspaceID`, `CreatedBy UserID`, `TotalAmount Money`, `Description string`, `Date time.Time`, `Fingerprint *string`, `Source TransactionSource`, `Type TransactionType`, `BankAccountID *BankAccountID`, `BankName string`, `Notes string`, `Entries []Entry`, `CreatedAt time.Time`, `UpdatedAt time.Time`

#### Scenario: Valid entries sum
- **WHEN** `ValidateEntries()` is called on a Transaction whose entries sum to the total amount
- **THEN** it SHALL return no error

#### Scenario: Entry sum mismatch
- **WHEN** `ValidateEntries()` is called on a Transaction whose entries do NOT sum to the total amount
- **THEN** it SHALL return `model.ErrEntrySumMismatch`

#### Scenario: No entries
- **WHEN** `ValidateEntries()` is called on a Transaction with zero entries
- **THEN** it SHALL return `model.ErrTransactionNoEntries`

#### Scenario: Transfer type constant
- **WHEN** `model.TypeTransfer` is referenced
- **THEN** it SHALL equal the string `"transfer"`

### Requirement: Transaction database migration adds transfer enum value
The database SHALL support `transfer` as a valid `transaction_type` enum value.

#### Scenario: Migration adds transfer value
- **WHEN** the migration is applied
- **THEN** `ALTER TYPE transaction_type ADD VALUE 'transfer'` SHALL be executed
- **AND** existing transactions SHALL not be affected

#### Scenario: Down migration is no-op
- **WHEN** the migration is rolled back
- **THEN** no action SHALL be taken (PostgreSQL does not support removing enum values)

### Requirement: Transaction type validation accepts transfer
The handler SHALL accept `transfer` as a valid transaction type in type filter query parameters and transaction creation/update requests.

#### Scenario: List with transfer type filter
- **WHEN** `GET /api/v1/workspaces/{id}/transactions?type=transfer` is called
- **THEN** only transactions with `type = 'transfer'` SHALL be returned

#### Scenario: Invalid type rejected
- **WHEN** a type filter value other than `expense`, `income`, or `transfer` is provided
- **THEN** the response SHALL be `400 Bad Request`
