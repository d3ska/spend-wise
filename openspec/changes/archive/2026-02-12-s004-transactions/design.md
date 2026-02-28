## Context

Workspaces, members, and categories are complete (`s003`). The project uses a flat package layout (`model/`, `store/`, `service/`, `handler/`), pgx+sqlc for database access, Chi router with JWT middleware, and `shopspring/decimal` for monetary precision. The permission system (`requireMembership`) is established in `WorkspaceService`.

Transactions are the core data entity of SpendWise — every expense is a transaction with one or more entries that split the amount across categories and optionally across participants. This is the first change that stores monetary values.

## Goals / Non-Goals

**Goals:**
- Transaction CRUD with atomic entry persistence (transaction + entries in one DB transaction)
- Flexible entry splitting: one transaction total split across N entries (category + optional participant)
- Entry-sum validation: entry amounts must exactly equal transaction total (domain invariant)
- Date-range filtering with pagination for listing
- Bank import with fingerprint-based deduplication
- All monetary values use `NUMERIC(19,4)` / `shopspring/decimal`

**Non-Goals:**
- No CSV parsing logic — the import endpoint accepts structured JSON for now, CSV parsing can be added later
- No transaction editing — create and delete only (immutable financial records)
- No recurring transactions
- No attachment/receipt uploads

## Decisions

### 1. Atomic create via pgx transaction, not sqlc

Transaction creation must insert one `transactions` row + N `entries` rows atomically. sqlc doesn't support multi-statement transactions natively. The store will use `pgxpool.Pool.Begin()` directly, wrapping manual inserts in `Begin`/`Commit` with `defer Rollback`.

**Alternative considered:** Two separate sqlc calls without a DB transaction. Risks orphaned transactions with no entries if the second call fails.

### 2. Entry-sum validation in the domain model, not the database

`Transaction.ValidateEntries()` checks that `SUM(entry.Amount) == transaction.TotalAmount` before persisting. This gives clear error messages (`ErrEntrySumMismatch`) rather than relying on a database CHECK constraint which would surface as an opaque DB error.

**Alternative considered:** PostgreSQL trigger that validates sums on insert. Harder to test, opaque errors, and couples business logic to the database.

### 3. Fingerprint dedup via partial unique index

Bank imports generate a fingerprint hash from (date, amount, description). A partial unique index `WHERE fingerprint IS NOT NULL` prevents duplicates without constraining manual transactions (which have NULL fingerprint). The store checks `ExistsByFingerprint` before inserting.

**Alternative considered:** Application-level dedup only (query before insert). Race condition if two imports run concurrently. The partial index is the correct solution.

### 4. Separate TransactionStore, not extending WorkspaceStore

Transactions get their own `store/transaction_store.go` because:
- The atomic create logic with pgx transactions is complex enough to warrant isolation
- The store needs direct pool access (for `Begin()`), not just a `DBTX` interface
- Queries are numerous (insert tx, insert entries, get by ID, list, delete, fingerprint check)

**Alternative considered:** Adding transaction methods to WorkspaceStore. Would make that file too large and mix concerns.

### 5. Import returns created + skipped counts, no CSV parsing yet

The import endpoint accepts a JSON array of transaction inputs (same shape as manual create, but with a `fingerprint` field). This keeps the first iteration simple. CSV parsing can be layered on top as a future enhancement that transforms CSV → JSON inputs.

**Alternative considered:** Multipart file upload with CSV parsing. Premature — the structured import covers the core dedup logic; CSV format varies by bank.

## Risks / Trade-offs

- **[Risk] Large entry lists per transaction** → No practical limit on entries per transaction. Acceptable for household scale. If needed, a max-entries validation can be added.
- **[Trade-off] No transaction editing** → Immutable records are simpler and safer for financial data. Users delete and re-create if needed.
- **[Trade-off] JSON import, not CSV** → Simpler to implement, but requires frontend to transform bank exports. CSV parsing is a future enhancement.
- **[Risk] Decimal precision in JSON** → Money amounts are serialized as strings in JSON to avoid floating-point precision loss. The handler must parse string amounts to `shopspring/decimal`.
