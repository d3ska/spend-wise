## ADDED Requirements

### Requirement: Background sync scheduler
The application SHALL run a background cron job that periodically syncs transactions from all active bank connections.

#### Scenario: Cron schedule
- **WHEN** the application starts
- **THEN** it SHALL start a cron scheduler that runs the bank sync job every 6 hours

#### Scenario: Graceful shutdown
- **WHEN** the application receives a shutdown signal
- **THEN** it SHALL stop the cron scheduler and wait for any in-progress sync to complete

#### Scenario: Skip expired connections
- **WHEN** the sync job runs
- **THEN** it SHALL only process bank connections with `status = 'active'` and `auth_expires_at > now()`

#### Scenario: Mark expired connections
- **WHEN** the sync job encounters a connection where `auth_expires_at <= now()`
- **THEN** it SHALL update its status to `expired`

### Requirement: Transaction sync service
The `service` package SHALL provide a `SyncService` that fetches transactions from GoCardless and inserts them into linked workspaces.

#### Scenario: Sync single bank account
- **WHEN** `SyncBankAccount(ctx, bankAccount)` is called
- **THEN** it SHALL fetch transactions from GoCardless using `date_from = last_synced_at - 3 days` and `date_to = today`
- **AND** it SHALL insert transactions into every workspace linked to that bank account

#### Scenario: First sync uses 3-month history
- **WHEN** `SyncBankAccount(ctx, bankAccount)` is called and `last_synced_at` is NULL
- **THEN** it SHALL use `date_from = today - 3 months`

#### Scenario: Deduplicate via fingerprint
- **WHEN** a transaction from GoCardless has an `internalTransactionId` that already exists as a fingerprint in the workspace
- **THEN** it SHALL be skipped (ON CONFLICT DO NOTHING on the fingerprint partial unique index)

#### Scenario: Map GoCardless transaction to SpendWise transaction
- **WHEN** a GoCardless booked transaction is mapped
- **THEN** it SHALL create a `Transaction` with: `Source = "bank"`, `Fingerprint = internalTransactionId`, `Description` = `creditorName` or `debtorName` or `remittanceInformation` (first non-empty), `Date = bookingDate`, `TotalAmount` from `transactionAmount`, `BankAccountID` set to the source bank account, `CreatedBy` set to the bank connection's owner user ID

#### Scenario: Bank transaction has single uncategorized entry
- **WHEN** a bank-synced transaction is created
- **THEN** it SHALL have exactly one `Entry` with the workspace's "Uncategorized" category, the full transaction amount, and no participant

#### Scenario: Uncategorized category resolution
- **WHEN** the sync processes a workspace that has no "Uncategorized" category
- **THEN** it SHALL create an "Uncategorized" category in that workspace before inserting transactions

#### Scenario: Update last_synced_at after successful sync
- **WHEN** all transactions for a bank account have been processed successfully
- **THEN** it SHALL update `bank_accounts.last_synced_at` to the current timestamp

#### Scenario: Sync failure does not crash the scheduler
- **WHEN** syncing a single bank account fails (GoCardless API error, DB error)
- **THEN** it SHALL log the error and continue syncing the remaining bank accounts
- **AND** it SHALL NOT update `last_synced_at` for the failed account

#### Scenario: Fetch once, insert to multiple workspaces
- **WHEN** a bank account is linked to multiple workspaces
- **THEN** the sync SHALL fetch transactions from GoCardless once and insert them into each linked workspace separately (with per-workspace fingerprint dedup)

### Requirement: Manual sync trigger endpoint
The `handler` package SHALL expose an endpoint to trigger an immediate sync for a specific bank account.

#### Scenario: Trigger sync endpoint
- **WHEN** `POST /api/v1/bank-connections/{id}/sync` is called
- **THEN** it SHALL trigger an immediate sync for all bank accounts under that connection
- **AND** it SHALL respond with `202 Accepted`

#### Scenario: Sync trigger requires ownership
- **WHEN** a user triggers sync for a connection they do not own
- **THEN** it SHALL respond with `404 Not Found`

#### Scenario: Sync trigger for expired connection
- **WHEN** a user triggers sync for an expired connection
- **THEN** it SHALL respond with `409 Conflict` and a message indicating re-authentication is required
