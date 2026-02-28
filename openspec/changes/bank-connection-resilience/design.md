## Context

Bank connections are deleted via `DELETE /bank-connections/{id}`, which CASCADE-deletes all `bank_accounts` rows. Transactions get `bank_account_id = NULL` (ON DELETE SET NULL). On reconnect, Enable Banking returns new `external_id` values (UIDs), so new `bank_accounts` rows are created. Fingerprint dedup prevents duplicate transactions, but old transactions remain orphaned — they can never be re-linked to the new bank account row. The 3-month sync lookback window makes re-linking via sync impossible for older transactions.

Current schema chain: `bank_connections` → (CASCADE) → `bank_accounts` → (SET NULL) → `transactions.bank_account_id`

## Goals / Non-Goals

**Goals:**
- Preserve transaction-to-account linkage across connection deletion and reconnection cycles
- Provide non-destructive deactivate/reactivate for bank connections
- Reuse existing bank account rows on reconnect to maintain referential integrity
- Store a permanent source account identifier (IBAN) on transactions

**Non-Goals:**
- Changing the hard-delete behavior (it stays, just not the default action)
- Frontend changes for deactivate/reactivate UI (backend-only in this change)

## Decisions

### 1. Store IBAN directly on transactions
**Decision**: Add `iban TEXT` column to `transactions` table.
**Rationale**: IBAN is the one stable identifier that doesn't change across reconnections. Together with the existing `total_currency`, the pair `(iban, currency)` uniquely identifies the source account. Unlike `bank_account_id`, it survives row deletion.
**Alternative considered**: Store only `bank_account_id` and prevent deletion — rejected because hard-delete is still a valid user action, and IBAN provides value even without an active bank account row (display, filtering, grouping).

### 2. Match by IBAN + currency + user_id on reconnect
**Decision**: In `Complete()`, before creating a new `bank_accounts` row, look up existing rows by `(iban, currency)` scoped to the user via join on `bank_connections.user_id`.
**Rationale**: IBAN alone isn't unique — Revolut multi-currency accounts share the same IBAN with different currencies. Adding user scoping prevents cross-user account reuse.
**Alternative considered**: Match by `external_id` — rejected because Enable Banking assigns new UIDs on each authorization, so `external_id` changes on reconnect.

### 3. Soft-delete via status field (inactive)
**Decision**: Add `BankConnectionInactive` status constant. `SyncAll` already filters by `status = 'active'`, so inactive connections are automatically skipped.
**Rationale**: No schema change needed — `status` is already `TEXT`. Minimal code change. Preserves all relationships (accounts, workspace links, transactions).
**Alternative considered**: Separate `is_active` boolean column — rejected as redundant; the existing status field already supports multiple states.

### 4. Keep ON DELETE SET NULL for hard delete
**Decision**: Hard-delete continues to NULL out `bank_account_id` on transactions. Transactions are preserved with their `iban` field intact.
**Rationale**: Transaction data is valuable for budgeting, reports, and tax records. Most fintech products (YNAB, Mint) keep transactions when bank links are removed.

### 5. Bulk re-link orphaned transactions on reconnect
**Decision**: After creating or reusing a bank account in `SyncBankAccount`, run a bulk UPDATE to re-link orphaned transactions by matching `iban + total_currency + workspace_id` where `bank_account_id IS NULL`.
**Rationale**: After a hard delete, transactions have `bank_account_id = NULL` but retain their `iban`. On reconnect, we know the new bank account's IBAN and currency. A single UPDATE re-links all orphaned transactions regardless of age — not limited by the 3-month sync lookback window.
**Query**: `UPDATE transactions SET bank_account_id = $1 WHERE iban = $2 AND total_currency = $3 AND bank_account_id IS NULL AND workspace_id = $4`

## Risks / Trade-offs

- **[Risk] IBAN is NULL for some accounts** → Reconnect-reuse is skipped; a new `bank_accounts` row is created. Acceptable — IBAN is populated for the vast majority of accounts.
- **[Risk] `external_id` UNIQUE constraint on update** → When reusing an existing bank account, the new `external_id` could theoretically conflict with another row. Mitigated by the fact that Enable Banking generates fresh UUIDs per authorization.
- **[Risk] Backfill won't cover already-orphaned transactions** → Transactions where `bank_account_id` is already NULL (from prior deletions) won't get `iban` backfilled via migration. Mitigation: on reconnect, the bulk re-link matches by IBAN+currency, so these transactions will be re-linked automatically if the user reconnects the same bank. For any remaining orphans, a one-time manual SQL fix can be applied.

## Migration Plan

1. Run migration `000011_bank_connection_resilience` — adds `iban` column, backfills from existing links, adds index
2. Deploy new backend code — new endpoints, updated sync/complete flows
3. One-time SQL to fix already-orphaned transactions (manual, per-deployment)
4. Rollback: run down migration (drops `iban` column and index), revert code
