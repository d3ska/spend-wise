## Why

When a user deletes a bank connection, `bank_accounts` rows are CASCADE-deleted and transactions lose their `bank_account_id` (ON DELETE SET NULL). On reconnect, new bank account rows are created with new IDs, and old transactions cannot be re-linked — they are orphaned with no way to identify their source account. Users need a non-destructive way to temporarily disable connections and a reliable way to reconnect without losing transaction linkage.

## What Changes

- Add `iban` column to `transactions` table as a permanent source account identifier that survives bank account deletion
- Add `inactive` status to bank connections for soft-delete (deactivate/reactivate) — stops syncing but preserves all links
- Add `POST /bank-connections/{id}/deactivate` and `POST /bank-connections/{id}/reactivate` endpoints
- On reconnect (Complete flow), reuse existing `bank_accounts` rows by matching IBAN + currency + user instead of always creating new rows
- Backfill `iban` on existing bank-synced transactions from their current `bank_account_id` link
- Expose `iban` in transaction API responses

## Capabilities

### New Capabilities
- `bank-connection-lifecycle`: Covers deactivate/reactivate flow, IBAN-based account reuse on reconnect, and the `inactive` connection status

### Modified Capabilities
- `bank-import`: Add IBAN tracking on synced transactions — the sync flow now populates `iban` from the source bank account
- `transaction-ledger`: Transaction entity gains an `iban` field; schema adds `iban TEXT` column to transactions table

## Impact

- **Database**: New migration adds `iban` column to `transactions`, adds index on `bank_accounts(iban, currency)`, backfills existing data
- **API**: Two new endpoints (deactivate/reactivate), `iban` field added to transaction responses
- **Service layer**: `bank_connection_service.Complete()` gains IBAN-based account reuse logic; `bank_sync_service.mapTransaction()` populates IBAN; new `Deactivate`/`Reactivate` methods
- **Security**: IBAN reconnect lookup scoped by `user_id` (joins through `bank_connections`) to prevent cross-user account reuse
