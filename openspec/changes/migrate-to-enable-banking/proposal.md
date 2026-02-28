## Why

GoCardless has discontinued free access to its Bank Account Data API for new personal users, making it unusable for our bank sync feature. Enable Banking offers a free tier for personal use (own accounts via restricted production app), supports all major Polish banks and Revolut, and provides up to 180-day consent duration (vs GoCardless's 90 days).

## What Changes

- **Replace GoCardless client** with Enable Banking client in the `banksync` package — new auth mechanism (JWT/RSA256 instead of API key/secret), new endpoints, new data models
- **Change auth flow**: GoCardless uses agreement → requisition → accounts. Enable Banking uses `POST /auth` → bank redirect → `POST /sessions` → accounts. The `requisition_id` column becomes `session_id`.
- **Update config**: Replace `GOCARDLESS_SECRET_ID`/`GOCARDLESS_SECRET_KEY` env vars with `ENABLEBANKING_APPLICATION_ID` and `ENABLEBANKING_KEY_PATH` (path to RSA private key PEM file)
- **Update institution lookup**: Enable Banking uses ASPSP name+country (not an institution ID). The `institution_id` field semantics change to store the ASPSP name.
- **Adapt transaction mapping**: Enable Banking returns different field names (`transaction_id` instead of `internalTransactionId`, `remittance_information` as array instead of string, `transaction_amount` with same shape). **BREAKING**: Fingerprint field changes — `transaction_id` is not guaranteed unique across all banks, so dedup strategy needs to combine multiple fields.
- **Update consent expiry**: Use per-ASPSP `maximum_consent_validity` (up to 180 days) instead of hardcoded 90 days.
- **DB migration**: Rename `requisition_id` column to `session_id` in `bank_connections` table.
- **Complete flow changes**: `POST /sessions` returns accounts with `uid` (used for data access) and `account_id` (IBAN). We can now populate the IBAN field directly.

## Capabilities

### New Capabilities
- `enable-banking-client`: JWT/RSA256 authenticated HTTP client for Enable Banking API (replaces GoCardless client). Covers token generation, ASPSP listing, auth initiation, session creation, and transaction fetching.

### Modified Capabilities
- `bank-connection`: Auth flow changes from GoCardless agreement/requisition model to Enable Banking auth/session model. Session ID replaces requisition ID. Consent duration becomes dynamic per-ASPSP.
- `bank-transaction-sync`: Transaction field mapping changes (different field names, array remittance info, composite fingerprint for dedup).

## Impact

- **Backend**: `banksync/` package rewritten, `service/bank_connection_service.go` and `service/bank_sync_service.go` updated, `config/config.go` updated, new DB migration
- **Config/Ops**: New env vars (`ENABLEBANKING_APPLICATION_ID`, `ENABLEBANKING_KEY_PATH`), RSA key file must be provisioned. Old GoCardless env vars removed.
- **Database**: Migration to rename `requisition_id` → `session_id`
- **Frontend**: No changes needed — the API contract between frontend and backend stays the same (same handler endpoints, same JSON shapes)
- **Dependencies**: No new Go dependencies needed (JWT signing uses stdlib `crypto/rsa`, `crypto/x509`)
