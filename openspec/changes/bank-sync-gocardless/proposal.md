## Why

Users currently add transactions manually or via CSV import. Direct bank integration via the GoCardless Bank Account Data API (PSD2, free tier) eliminates most manual entry by automatically syncing booked transactions from connected European bank accounts on a recurring schedule.

## What Changes

- Add bank connection management: users authenticate with their bank via GoCardless OAuth flow (one-time per bank, reusable across workspaces)
- Add bank account linking: users choose which bank accounts feed into which workspaces
- Add background cron sync: fetch new transactions every ~6 hours using date-range filtering (`date_from`/`date_to`) with fingerprint deduplication
- Add new transaction source `"bank"` alongside existing `"manual"` and `"import"`
- Store bank origin metadata on transactions (`bank_account_id` FK)
- Bank-synced transactions land with a default "Uncategorized" category (no review gate)
- Import 3 months of history on first connection
- Track 90-day auth expiry per connection with tiered UI notifications (settings countdown, 14-day warning banner, expired modal)
- Support multiple banks per user, multiple bank accounts per workspace

## Capabilities

### New Capabilities
- `bank-connection`: Managing GoCardless bank connections (create requisition, OAuth redirect, store credentials, track expiry, reconnect flow)
- `bank-account-linking`: Linking bank accounts to workspaces via join table, selecting which accounts sync to which workspaces
- `bank-transaction-sync`: Background cron job that fetches transactions from GoCardless, maps them to the SpendWise transaction model, deduplicates via fingerprint, and inserts with source `"bank"`
- `bank-auth-expiry-notifications`: Frontend notification system for 90-day re-auth (settings page status, 14-day warning banner, expired-state modal)

### Modified Capabilities
- `transaction-ledger`: Add `"bank"` to the `TransactionSource` enum; add nullable `bank_account_id` FK to transactions table
- `bank-import`: Extend fingerprint deduplication to cover bank-synced transactions (source `"bank"` uses `internalTransactionId` from GoCardless as fingerprint)

## Impact

- **Database**: 3 new tables (`bank_connections`, `bank_accounts`, `workspace_bank_accounts`), 1 altered table (`transactions` — new column + enum value), 2 new migrations
- **Backend**: New `banksync` package (GoCardless HTTP client), new service + handler + store layers for bank connections/accounts, cron scheduler infrastructure
- **Frontend**: Bank connection UI in workspace settings, bank selection flow, account-workspace linking, auth expiry notifications, bank origin badge on transactions
- **Dependencies**: GoCardless API v2 (external HTTP dependency), cron/scheduler library (e.g. `robfig/cron`)
- **Config**: New env vars for GoCardless `secret_id` and `secret_key`
- **APIs**: New endpoints for bank connection CRUD, institution listing, OAuth callback, manual sync trigger
