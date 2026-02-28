## 1. Database Migration

- [x] 1.1 Create `backend/migrations/000011_bank_connection_resilience.up.sql` — add `iban TEXT` column to transactions, backfill from bank_accounts, add `idx_bank_accounts_iban_currency` index
- [x] 1.2 Create `backend/migrations/000011_bank_connection_resilience.down.sql` — drop index and column
- [x] 1.3 Run `make migrate-up` to apply migration

## 2. Model Changes

- [x] 2.1 Add `BankConnectionInactive` status constant to `backend/model/bank_connection.go`
- [x] 2.2 Add `IBAN *string` field to `Transaction` struct in `backend/model/transaction.go`

## 3. SQL Queries

- [x] 3.1 Update `backend/db/queries/transactions.sql` — add `iban` to INSERT, SELECT, and RETURNING clauses in all transaction queries
- [x] 3.2 Add `GetBankAccountByIBANAndCurrencyForUser` query to `backend/db/queries/bank_accounts.sql` — lookup by IBAN + currency + user_id (join bank_connections)
- [x] 3.3 Add `UpdateBankAccountConnection` query to `backend/db/queries/bank_accounts.sql` — update external_id and bank_connection_id
- [x] 3.4 Add `RelinkOrphanedTransactions` query to `backend/db/queries/transactions.sql` — `UPDATE transactions SET bank_account_id = $1 WHERE iban = $2 AND total_currency = $3 AND bank_account_id IS NULL AND workspace_id = $4`
- [x] 3.5 Run `make sqlc` to regenerate store code

## 4. Store Layer

- [x] 4.1 Update `backend/store/transaction_store.go` — pass `IBAN` in InsertTransactionParams, map `IBAN` in all `toModelTransaction*` functions
- [x] 4.2 Add `GetByIBANAndCurrencyForUser(ctx, iban, currency, userID)` method to `backend/store/bank_account_store.go`
- [x] 4.3 Add `UpdateConnection(ctx, id, externalID, connID)` method to `backend/store/bank_account_store.go`
- [x] 4.4 Add `RelinkOrphanedTransactions(ctx, bankAccountID, iban, currency, workspaceID)` method to `backend/store/transaction_store.go`

## 5. Service Layer

- [x] 5.1 Update `mapTransaction()` in `backend/service/bank_sync_service.go` — accept `bankIBAN *string` param and set `IBAN` on returned transaction
- [x] 5.2 Update `SyncBankAccount` call site to pass `acct.IBAN` to `mapTransaction()`
- [x] 5.6 Add bulk re-link step in `SyncBankAccount` — for each workspace, call `RelinkOrphanedTransactions` with the bank account's IBAN and currency before processing new transactions
- [x] 5.3 Add `Deactivate(ctx, userID, connID)` method to `backend/service/bank_connection_service.go`
- [x] 5.4 Add `Reactivate(ctx, userID, connID)` method to `backend/service/bank_connection_service.go`
- [x] 5.5 Update `Complete()` in `backend/service/bank_connection_service.go` — before creating a new bank account, check for existing by IBAN + currency + user; if found, update its external_id and bank_connection_id

## 6. Handler + Router

- [x] 6.1 Add `DeactivateConnection` handler to `backend/handler/handler_bank.go` — `POST /bank-connections/{id}/deactivate`
- [x] 6.2 Add `ReactivateConnection` handler to `backend/handler/handler_bank.go` — `POST /bank-connections/{id}/reactivate`
- [x] 6.3 Register new routes in `backend/handler/router.go`
- [x] 6.4 Expose `iban` field in transaction API responses in `backend/handler/handler_transaction.go`

## 7. Verification

- [x] 7.1 Run `make build` to verify compilation
- [x] 7.2 Run `make test` to verify all tests pass
- [ ] 7.3 Manual test: sync transactions and verify `iban` is populated
- [ ] 7.4 Manual test: deactivate connection, verify sync skips it, reactivate, verify sync resumes
- [ ] 7.5 Manual test: hard delete connection, reconnect same bank, verify orphaned transactions are re-linked via IBAN+currency
