## 1. Database Migrations

- [x] 1.1 Create migration for `bank_connections` table (id, user_id FK, institution_id, institution_name, requisition_id UNIQUE, status DEFAULT 'active', auth_expires_at, created_at, updated_at)
- [x] 1.2 Create migration for `bank_accounts` table (id, bank_connection_id FK CASCADE, external_id UNIQUE, iban nullable, name, last_synced_at nullable, created_at, updated_at)
- [x] 1.3 Create migration for `workspace_bank_accounts` join table (workspace_id FK CASCADE, bank_account_id FK CASCADE, UNIQUE constraint)
- [x] 1.4 Create migration to add `'bank'` to `transaction_source` enum and add `bank_account_id` (BIGINT FK to bank_accounts ON DELETE SET NULL) column to `transactions` table
- [x] 1.5 Create down migrations for all of the above

## 2. Model Layer

- [x] 2.1 Define `BankConnectionID`, `BankAccountID` typed IDs and `BankConnectionStatus` enum (active/expired/revoked) in model package
- [x] 2.2 Define `BankConnection` struct with all fields (ID, UserID, InstitutionID, InstitutionName, RequisitionID, Status, AuthExpiresAt, CreatedAt, UpdatedAt)
- [x] 2.3 Define `BankAccount` struct with all fields (ID, BankConnectionID, ExternalID, IBAN, Name, LastSyncedAt, CreatedAt, UpdatedAt)
- [x] 2.4 Add `BankAccountID *BankAccountID` field to existing `Transaction` struct
- [x] 2.5 Add `"bank"` to `TransactionSource` enum constants
- [x] 2.6 Add sentinel errors: `ErrBankConnectionNotFound`, `ErrBankAccountNotFound`, `ErrBankAccountAlreadyExists`, `ErrBankConnectionExpired`

## 3. GoCardless Client (`banksync` package)

- [x] 3.1 Create `backend/banksync/` package with `Client` interface (GetInstitutions, CreateAgreement, CreateRequisition, GetRequisition, GetAccountTransactions)
- [x] 3.2 Define GoCardless response types: `Institution`, `Agreement`, `Requisition`, `Transaction`, `Amount`
- [x] 3.3 Implement HTTP client with token management (POST /api/v2/token/new/, POST /api/v2/token/refresh/) including auto-refresh on 401
- [x] 3.4 Implement `GetInstitutions(ctx, countryCode)` — GET /api/v2/institutions/?country={code}
- [x] 3.5 Implement `CreateAgreement(ctx, institutionID)` — POST /api/v2/agreements/enduser/
- [x] 3.6 Implement `CreateRequisition(ctx, institutionID, redirectURL, agreementID)` — POST /api/v2/requisitions/
- [x] 3.7 Implement `GetRequisition(ctx, requisitionID)` — GET /api/v2/requisitions/{id}/
- [x] 3.8 Implement `GetAccountTransactions(ctx, accountID, dateFrom, dateTo)` — GET /api/v2/accounts/{id}/transactions/?date_from=...&date_to=... (return booked only)

## 4. Store Layer (sqlc queries + store wrappers)

- [x] 4.1 Write sqlc queries for bank_connections CRUD (Create, GetByID, ListByUser, ListActive, UpdateStatus, Delete)
- [x] 4.2 Write sqlc queries for bank_accounts (Create, ListByConnection, ListByWorkspace with JOIN, UpdateLastSyncedAt)
- [x] 4.3 Write sqlc queries for workspace_bank_accounts (Link ON CONFLICT DO NOTHING, Unlink, ListWorkspacesByBankAccount)
- [x] 4.4 Run `make sqlc` to generate Go code from new queries
- [x] 4.5 Create `BankConnectionStore` with model mapping (pgtype conversions, error wrapping)
- [x] 4.6 Create `BankAccountStore` with model mapping and workspace linking methods
- [x] 4.7 Update `TransactionStore` to handle `bank_account_id` column in Create and query results

## 5. Service Layer

- [x] 5.1 Create `BankConnectionService` with Initiate (create agreement + requisition + persist), Complete (fetch accounts + persist), Delete, Reconnect, ListByUser, ListInstitutions
- [x] 5.2 Create `BankAccountLinkingService` with LinkToWorkspace, UnlinkFromWorkspace, ListByWorkspace (all with permission checks via workspace membership)
- [x] 5.3 Create `BankSyncService` with SyncAll (iterate active connections), SyncBankAccount (fetch txns, map, dedup, insert per workspace), EnsureUncategorizedCategory
- [x] 5.4 Implement transaction mapping: GoCardless transaction → SpendWise Transaction with single "Uncategorized" entry, source="bank", fingerprint=internalTransactionId
- [x] 5.5 Implement date-range logic: first sync = today - 3 months, subsequent = last_synced_at - 3 days

## 6. Handler Layer (HTTP endpoints)

- [x] 6.1 Create `BankHandler` with ListInstitutions (GET /api/v1/banks?country=)
- [x] 6.2 Add InitiateConnection (POST /api/v1/bank-connections) — returns connection ID + auth link
- [x] 6.3 Add CompleteConnection (POST /api/v1/bank-connections/{id}/complete) — returns discovered accounts
- [x] 6.4 Add ListConnections (GET /api/v1/bank-connections) — returns user's connections with expiry status
- [x] 6.5 Add DeleteConnection (DELETE /api/v1/bank-connections/{id})
- [x] 6.6 Add ReconnectConnection (POST /api/v1/bank-connections/{id}/reconnect) — returns new auth link
- [x] 6.7 Add TriggerSync (POST /api/v1/bank-connections/{id}/sync) — 202 Accepted
- [x] 6.8 Add LinkBankAccount (POST /api/v1/workspaces/{id}/bank-accounts)
- [x] 6.9 Add UnlinkBankAccount (DELETE /api/v1/workspaces/{id}/bank-accounts/{bankAccountId})
- [x] 6.10 Add ListLinkedBankAccounts (GET /api/v1/workspaces/{id}/bank-accounts)
- [x] 6.11 Add computed `days_remaining` and `expiry_status` fields to bank connection JSON responses

## 7. Router & Composition Root

- [x] 7.1 Register all new bank routes in router.go (under protected group)
- [x] 7.2 Add `GoCardless` config section to config.go (SecretID, SecretKey, BaseURL env vars)
- [x] 7.3 Add `robfig/cron/v3` dependency via `go get`
- [x] 7.4 Wire up BankConnectionStore, BankAccountStore, GoCardless Client, BankConnectionService, BankAccountLinkingService, BankSyncService, BankHandler in cmd/api/main.go
- [x] 7.5 Initialize cron scheduler in main.go, register SyncAll job at "0 */6 * * *", start alongside HTTP server
- [x] 7.6 Add cron scheduler stop to graceful shutdown sequence (stop cron before shutting down HTTP server)

## 8. Frontend — Bank Connection Flow

- [x] 8.1 Add TypeScript types for BankConnection, BankAccount, Institution, and API response shapes
- [x] 8.2 Create API hooks: useInstitutions, useInitiateConnection, useCompleteConnection, useListConnections, useDeleteConnection, useReconnectConnection
- [x] 8.3 Create API hooks: useLinkedBankAccounts, useLinkBankAccount, useUnlinkBankAccount, useTriggerSync
- [x] 8.4 Build bank selection UI component (country dropdown + institution list with logos)
- [x] 8.5 Build bank auth redirect flow (initiate → redirect to bank → callback page that calls complete)
- [x] 8.6 Build bank account linking UI in workspace settings (list discovered accounts, toggle which link to this workspace)

## 9. Frontend — Bank Connection Management & Notifications

- [x] 9.1 Add bank connections section to WorkspaceSettingsPage showing linked accounts with institution name, IBAN, last synced, expiry date, and status indicator (green/yellow/red)
- [x] 9.2 Add "Reconnect" button for warning/expired connections that triggers the reconnect flow
- [x] 9.3 Add "Disconnect" button to unlink bank accounts from workspace
- [x] 9.4 Add yellow warning banner in Sidebar when any linked connection has expiry_status="warning"
- [x] 9.5 Add expired connection modal on DashboardPage when any linked connection has expiry_status="expired" (dismissible per session)
- [x] 9.6 Add bank origin badge on transaction list items for source="bank" transactions (show institution name)
