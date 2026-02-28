## ADDED Requirements

### Requirement: Bank connection deactivation
The system SHALL allow a user to deactivate a bank connection, setting its status to `inactive`. Inactive connections SHALL be skipped during transaction sync but SHALL preserve all linked bank accounts, workspace links, and transactions.

#### Scenario: Deactivate an active connection
- **WHEN** `POST /api/v1/bank-connections/{id}/deactivate` is called by the connection owner
- **THEN** the connection status SHALL be set to `inactive`
- **AND** the response SHALL be `204 No Content`

#### Scenario: Deactivate skips sync
- **WHEN** `SyncAll` runs and a connection has status `inactive`
- **THEN** the connection SHALL be skipped
- **AND** no transactions SHALL be fetched for its accounts

#### Scenario: Deactivate preserves data
- **WHEN** a connection is deactivated
- **THEN** all `bank_accounts` rows SHALL remain
- **AND** all `workspace_bank_accounts` links SHALL remain
- **AND** all transactions with `bank_account_id` referencing those accounts SHALL remain unchanged

#### Scenario: Deactivate by non-owner
- **WHEN** a user who does not own the connection attempts to deactivate it
- **THEN** the response SHALL be `404 Not Found`

### Requirement: Bank connection reactivation
The system SHALL allow a user to reactivate an inactive bank connection, setting its status back to `active`.

#### Scenario: Reactivate an inactive connection
- **WHEN** `POST /api/v1/bank-connections/{id}/reactivate` is called by the connection owner
- **AND** the connection status is `inactive`
- **THEN** the connection status SHALL be set to `active`
- **AND** the response SHALL be `204 No Content`

#### Scenario: Reactivate a non-inactive connection
- **WHEN** `POST /api/v1/bank-connections/{id}/reactivate` is called
- **AND** the connection status is not `inactive`
- **THEN** the response SHALL be an error indicating the connection is not inactive

#### Scenario: Reactivate by non-owner
- **WHEN** a user who does not own the connection attempts to reactivate it
- **THEN** the response SHALL be `404 Not Found`

### Requirement: Bank account reuse on reconnect by IBAN
The system SHALL reuse existing `bank_accounts` rows when a user reconnects a bank and the Enable Banking session returns an account with an IBAN + currency pair that matches an existing bank account owned by the same user.

#### Scenario: Reconnect reuses existing bank account
- **WHEN** a bank connection `Complete` flow processes an account with IBAN `PL123` and currency `PLN`
- **AND** an existing `bank_accounts` row with the same IBAN and currency exists for the same user
- **THEN** the existing row's `external_id` and `bank_connection_id` SHALL be updated to the new values
- **AND** no new `bank_accounts` row SHALL be created
- **AND** existing transactions linked to that bank account SHALL remain linked

#### Scenario: Reconnect creates new account when no IBAN match
- **WHEN** a bank connection `Complete` flow processes an account with an IBAN that does not match any existing bank account for the user
- **THEN** a new `bank_accounts` row SHALL be created

#### Scenario: Reconnect creates new account when IBAN is NULL
- **WHEN** a bank connection `Complete` flow processes an account with no IBAN
- **THEN** a new `bank_accounts` row SHALL be created

#### Scenario: IBAN match is scoped by user
- **WHEN** user A reconnects with IBAN `PL123` and currency `PLN`
- **AND** user B has an existing bank account with the same IBAN and currency
- **THEN** user B's bank account SHALL NOT be reused
- **AND** a new bank account SHALL be created for user A

### Requirement: Bulk re-link orphaned transactions on reconnect
The system SHALL re-link orphaned transactions (where `bank_account_id IS NULL`) to a bank account when syncing, by matching `iban + total_currency + workspace_id`.

#### Scenario: Re-link after hard delete and reconnect
- **WHEN** a bank account is synced for a workspace
- **AND** orphaned transactions exist in that workspace with matching `iban` and `total_currency`
- **THEN** those transactions' `bank_account_id` SHALL be updated to the current bank account's ID

#### Scenario: Re-link does not affect already-linked transactions
- **WHEN** a bank account is synced for a workspace
- **AND** transactions exist with a non-NULL `bank_account_id`
- **THEN** those transactions SHALL NOT be modified

#### Scenario: Re-link matches by IBAN and currency
- **WHEN** orphaned transactions exist with IBAN `PL123` and currency `EUR`
- **AND** the syncing bank account has IBAN `PL123` and currency `PLN`
- **THEN** those transactions SHALL NOT be re-linked (currency mismatch)
