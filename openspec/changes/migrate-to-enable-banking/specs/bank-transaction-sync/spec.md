## MODIFIED Requirements

### Requirement: Transaction sync service
The `service` package SHALL provide a `SyncService` that fetches transactions from Enable Banking and inserts them into linked workspaces.

#### Scenario: Sync single bank account
- **WHEN** `SyncBankAccount(ctx, bankAccount)` is called
- **THEN** it SHALL fetch transactions from Enable Banking using `date_from = last_synced_at - 3 days` and `date_to = today`
- **AND** it SHALL insert transactions into every workspace linked to that bank account

#### Scenario: First sync uses 3-month history
- **WHEN** `SyncBankAccount(ctx, bankAccount)` is called and `last_synced_at` is NULL
- **THEN** it SHALL use `date_from = today - 3 months`

#### Scenario: Deduplicate via composite fingerprint
- **WHEN** a transaction from Enable Banking is processed
- **THEN** the fingerprint SHALL be computed as `sha256(transaction_id + "|" + booking_date + "|" + amount + "|" + currency)` when `transaction_id` is present
- **AND** the fingerprint SHALL be computed as `sha256(booking_date + "|" + amount + "|" + currency + "|" + first_remittance_info)` when `transaction_id` is absent
- **AND** if a transaction with the same fingerprint already exists in the workspace, it SHALL be skipped

#### Scenario: Map Enable Banking transaction to SpendWise transaction
- **WHEN** an Enable Banking booked transaction is mapped
- **THEN** it SHALL create a `Transaction` with: `Source = "bank"`, `Fingerprint` = composite hash (see dedup scenario), `Description` = first non-empty of `creditor_name`, `debtor_name`, or first element of `remittance_information`, `Date = booking_date`, `TotalAmount` from `transaction_amount`, `BankAccountID` set to the source bank account, `CreatedBy` set to the bank connection's owner user ID

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
- **WHEN** syncing a single bank account fails (API error, DB error)
- **THEN** it SHALL log the error and continue syncing the remaining bank accounts
- **AND** it SHALL NOT update `last_synced_at` for the failed account

#### Scenario: Fetch once, insert to multiple workspaces
- **WHEN** a bank account is linked to multiple workspaces
- **THEN** the sync SHALL fetch transactions from Enable Banking once and insert them into each linked workspace separately (with per-workspace fingerprint dedup)

#### Scenario: Handle transaction pagination
- **WHEN** Enable Banking returns transactions with a `continuation_key`
- **THEN** the sync SHALL fetch all pages before processing transactions
