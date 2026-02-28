## Why

Transactions are the core of SpendWise — every expense, purchase, or payment is recorded as a transaction. The flexible splitting model (one transaction → multiple entries across categories and participants) is the key differentiator from simple expense trackers. Bank import with fingerprint-based deduplication prevents double-counting. This is the primary business value of the application.

**Depends on:** `workspaces-and-categories` (Workspace, Category entities, permission system)

## What Changes

- Create `Transaction` domain entity with typed `TransactionID`, `TransactionSource` (manual|import), nullable `Fingerprint` for dedup, `ValidateEntries()` method enforcing entries sum to total
- Create `Entry` domain entity with typed `EntryID`, category assignment, optional participant assignment, `Money` amount
- Create `domain.TransactionRepository` interface (Create [atomic: tx + entries], GetByID [with entries], ListByWorkspace [with date range + pagination], ExistsByFingerprint, Delete)
- Create database migration `000004_create_transactions` with:
  - `NUMERIC(19,4)` for all monetary columns
  - `transaction_source` PostgreSQL enum
  - Partial unique index on `(workspace_id, fingerprint) WHERE fingerprint IS NOT NULL` for dedup
  - Composite index on `(workspace_id, date)` for efficient date-range queries
  - `entries` table with FK to transactions (CASCADE delete), FK to categories, nullable FK to users (participant)
- Create sqlc queries for transactions and entries (`db/queries/transactions.sql`)
- Create `adapter/postgres/transaction_repo.go` with:
  - Atomic `Create` using pgx transaction: `Begin` → insert transaction → insert entries → `Commit`, with proper `defer Rollback` error handling (100go.co #54)
  - `GetByID` eagerly loads entries
  - `ListByWorkspace` supports `?from=&to=&limit=&offset=` filtering
- Create `app/transaction_service.go` with co-located DTOs:
  - `Create` — verifies editor+ permission, validates entry sum, persists atomically
  - `Import` — verifies permission, checks fingerprint for duplicates (returns `ErrDuplicateFingerprint` if exists), accepts `io.Reader` for CSV/bank statement parsing (100go.co #46), creates transaction with source=import
  - `List` — returns transactions for workspace in date range with pagination
  - `Get`, `Delete` — with permission checks
- Create `adapter/httpapi/handler_transaction.go` with endpoints:
  - `POST /workspaces/{id}/transactions` — create manual transaction with entries
  - `GET /workspaces/{id}/transactions?from=&to=&limit=&offset=` — list with filtering
  - `POST /workspaces/{id}/transactions/import` — bank statement import
  - `GET /workspaces/{id}/transactions/{txID}` — get with entries
  - `DELETE /workspaces/{id}/transactions/{txID}` — delete
- Wire transaction service/handler into router

## Capabilities

### New Capabilities
- `transaction-ledger`: Transaction CRUD with flexible entry splitting (one transaction → N entries across categories/participants), entry-sum validation (entries must equal total), date-range filtering with pagination, atomic persistence of transaction + entries
- `bank-import`: Bank statement import with fingerprint-based deduplication via partial unique index, `io.Reader` interface for source-agnostic parsing (HTTP upload, file, stream)

### Modified Capabilities
- `http-server`: Add transaction routes under `/api/v1/workspaces/{id}/transactions`

## Impact

- **API:** 5 new endpoints for transaction CRUD + import
- **Database:** Migration 004 creates `transactions` and `entries` tables — the largest and most complex migration (partial unique index, composite index, two tables)
- **Domain invariant:** Entry amounts must sum to transaction total — enforced in domain layer via `ValidateEntries()`, not just in the database
- **Code:** ~4 new files (1 domain, 1 repo, 1 service, 1 handler)
- **Financial data:** First change that stores monetary values. All amounts use `NUMERIC(19,4)` in DB and `shopspring/decimal` in Go — never floating-point.
