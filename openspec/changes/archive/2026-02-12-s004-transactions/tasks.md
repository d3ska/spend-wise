## 1. Model Entities

- [x] 1.1 Create `model/transaction.go` — `TransactionID` typed wrapper, `EntryID` typed wrapper, `TransactionSource` (SourceManual/SourceImport), `Transaction` struct (ID, WorkspaceID, CreatedBy, TotalAmount Money, Description, Date, Fingerprint *string, Source, Notes, Entries []Entry, CreatedAt, UpdatedAt), `ValidateEntries()` method returning `ErrTransactionNoEntries` or `ErrEntrySumMismatch`
- [x] 1.2 Add `Entry` struct to `model/transaction.go` — (ID, TransactionID, CategoryID, ParticipantID *UserID, Amount Money, Note)

## 2. Database Migration

- [x] 2.1 Create `migrations/000004_create_transactions.up.sql` — `transaction_source` enum, `transactions` table (BIGSERIAL PK, workspace_id FK ON DELETE CASCADE, created_by FK, total_amount NUMERIC(19,4), currency TEXT DEFAULT 'PLN', description, date DATE, fingerprint nullable, source, notes, timestamps), `entries` table (BIGSERIAL PK, transaction_id FK ON DELETE CASCADE, category_id FK, participant_id nullable FK, amount NUMERIC(19,4), currency, note), composite index `(workspace_id, date)`, partial unique index `(workspace_id, fingerprint) WHERE fingerprint IS NOT NULL`
- [x] 2.2 Create `migrations/000004_create_transactions.down.sql` — drop entries, transactions, transaction_source enum

## 3. Queries and Store

- [x] 3.1 Create `db/queries/transactions.sql` — sqlc queries: InsertTransaction, GetTransactionByID, ListTransactionsByWorkspace (with date range + limit + offset), DeleteTransaction, ExistsByFingerprint
- [x] 3.2 Create `db/queries/entries.sql` — sqlc queries: InsertEntry, ListEntriesByTransaction
- [x] 3.3 Run `sqlc generate` to update generated code in `store/`
- [x] 3.4 Create `store/transaction_store.go` — `TransactionStore` with pool dependency (for `Begin()`). Methods: Create (atomic: begin tx → insert transaction → insert entries → commit), GetByID (loads entries), ListByWorkspace (date range + pagination), Delete, ExistsByFingerprint. Model mapping with `shopspring/decimal` for Money fields.

## 4. Service Layer

- [x] 4.1 Create `service/transaction_service.go` — `TransactionService` struct with transaction store, workspace store dependencies. Co-located DTOs: CreateTransactionInput, CreateEntryInput, ImportTransactionsInput, ListTransactionsInput. Methods:
  - `Create` — validate entries sum, check editor+ permission, persist atomically
  - `GetTransaction` — viewer+ permission, return with entries
  - `ListTransactions` — viewer+ permission, date range + pagination
  - `DeleteTransaction` — editor+ permission
  - `ImportTransactions` — editor+ permission, check fingerprints, skip duplicates, return imported/skipped counts

## 5. HTTP Handler

- [x] 5.1 Create `handler/handler_transaction.go` — `TransactionHandler` struct wrapping service. Methods: CreateTransaction, ListTransactions, GetTransaction, DeleteTransaction, ImportTransactions. Parse workspace ID, transaction ID from URL params. Parse `from`, `to`, `limit`, `offset` from query params. Parse Money amounts from string.

## 6. Router Wiring

- [x] 6.1 Update `handler/router.go` — add `TransactionHandler` to `RouterConfig`, register routes under `/api/v1/workspaces/{id}`:
  - `POST /transactions`, `GET /transactions`, `POST /transactions/import`
  - `GET /transactions/{txID}`, `DELETE /transactions/{txID}`
- [x] 6.2 Update `cmd/api/main.go` — create TransactionStore, TransactionService, TransactionHandler, pass to NewRouter

## 7. Verification

- [x] 7.1 Run `go build ./...` — must compile with zero errors
- [x] 7.2 Run `go vet ./...` — must pass with zero warnings
- [x] 7.3 Run `go test -count=1 ./...` — all tests must pass
