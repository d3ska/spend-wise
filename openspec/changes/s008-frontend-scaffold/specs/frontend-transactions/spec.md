## ADDED Requirements

### Requirement: Transactions data table
The transactions page SHALL display a paginated data table with columns: Date, Description, Category, Amount, and actions.

#### Scenario: Table renders transactions
- **WHEN** the transactions page loads
- **THEN** it SHALL fetch and display transactions for the current workspace and date range

#### Scenario: Pagination
- **WHEN** there are more transactions than the page size (default 50)
- **THEN** pagination controls SHALL appear to navigate between pages

#### Scenario: Empty state
- **WHEN** there are no transactions in the selected range
- **THEN** the table SHALL show "No transactions found" with a prompt to create one

### Requirement: Create transaction form
The transactions page SHALL provide a form to create a new transaction with entries (split among participants).

#### Scenario: Create transaction
- **WHEN** the user fills in description, date, category, amount, and participant entries and submits
- **THEN** the app SHALL POST to the create transaction endpoint
- **AND** on success, the transaction list SHALL refresh

#### Scenario: Validation errors
- **WHEN** the user submits a form with missing required fields
- **THEN** inline validation errors SHALL be displayed (powered by Zod + React Hook Form)

### Requirement: Delete transaction
The transactions page SHALL allow deleting a transaction with confirmation.

#### Scenario: Delete with confirmation
- **WHEN** the user clicks delete on a transaction
- **THEN** a confirmation dialog SHALL appear
- **AND** on confirm, the app SHALL call the delete endpoint and refresh the list

### Requirement: Date range filter
The transactions page SHALL include a date range picker to filter transactions.

#### Scenario: Filter by date range
- **WHEN** the user selects a date range
- **THEN** the transactions list SHALL refetch with the new `from` and `to` parameters

### Requirement: Bank import
The transactions page SHALL provide a way to import transactions from a file.

#### Scenario: Import transactions
- **WHEN** the user uploads a file and submits
- **THEN** the app SHALL POST to the import endpoint
- **AND** display the count of imported and skipped transactions
