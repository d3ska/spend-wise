## 1. Database Migration

- [x] 1.1 Create migration `000012_rename_requisition_to_session.up.sql` — rename `requisition_id` column to `session_id`, drop old unique index, create new unique index on `session_id`
- [x] 1.2 Create migration `000012_rename_requisition_to_session.down.sql` — rename `session_id` back to `requisition_id`, restore original unique index

## 2. Config

- [x] 2.1 Replace `GoCardlessConfig` struct with `EnableBankingConfig` in `config/config.go` — fields: `ApplicationID` (env `ENABLEBANKING_APP_ID`), `KeyPath` (env `ENABLEBANKING_KEY_PATH`), `BaseURL` (env `ENABLEBANKING_BASE_URL`, default `https://api.enablebanking.com`)
- [x] 2.2 Update `Config` struct to reference `EnableBanking EnableBankingConfig` instead of `GoCardless GoCardlessConfig`

## 3. Model Layer

- [x] 3.1 Rename `RequisitionID` field to `SessionID` in `model.BankConnection` struct
- [x] 3.2 Update comment on `BankConnection` to reference Enable Banking instead of GoCardless

## 4. Enable Banking Client (banksync package rewrite)

- [x] 4.1 Rewrite `banksync/client.go` — replace `Client` interface methods with: `GetASPSPs(ctx, country) ([]ASPSP, error)`, `StartAuth(ctx, aspspName, country, redirectURL string, validUntil time.Time) (AuthResult, error)`, `CreateSession(ctx, code string) (Session, error)`, `GetAccountTransactions(ctx, accountUID string, dateFrom, dateTo time.Time) ([]Transaction, error)`
- [x] 4.2 Define new types: `ASPSP` (Name, Country, Logo, BIC, MaxConsentValidity), `AuthResult` (URL, AuthorizationID), `Session` (SessionID, Accounts []SessionAccount), `SessionAccount` (UID, AccountID, Name, Currency)
- [x] 4.3 Update `Transaction` struct — `TransactionID` replaces `InternalTransactionID`, `RemittanceInformation []string` (array), add `CreditDebitIndicator`, `Status` fields
- [x] 4.4 Rewrite `banksync/http_client.go` — replace GoCardless token auth with JWT/RSA256 generation: load RSA private key from PEM file, generate JWT with `{alg: RS256, kid: appID}` header and `{iss: enablebanking.com, aud: api.enablebanking.com}` payload (1-hour TTL)
- [x] 4.5 Implement `GetASPSPs` — `GET /aspsps?country={code}`, parse ASPSP list with `maximum_consent_validity`
- [x] 4.6 Implement `StartAuth` — `POST /auth` with `{access: {valid_until}, aspsp: {name, country}, state: uuid, redirect_url, psu_type: "personal"}`
- [x] 4.7 Implement `CreateSession` — `POST /sessions` with `{code}`, return session ID and accounts
- [x] 4.8 Implement `GetAccountTransactions` — `GET /accounts/{uid}/transactions?date_from=&date_to=`, handle `continuation_key` pagination, filter to `BOOK` status only

## 5. Store Layer

- [x] 5.1 Update sqlc queries in `db/queries/bank_connections.sql` — rename `requisition_id` references to `session_id`
- [x] 5.2 Rename `UpdateBankConnectionRequisition` query to `UpdateBankConnectionSession`
- [x] 5.3 Regenerate sqlc: `make sqlc`
- [x] 5.4 Update `store/bank_connection_store.go` — rename `RequisitionID` mappings to `SessionID`, rename `UpdateRequisition` method to `UpdateSession`

## 6. Service Layer

- [x] 6.1 Update `service/bank_connection_service.go` — replace `gcClient` references with Enable Banking client calls: `GetASPSPs` instead of `GetInstitutions`, `StartAuth` instead of `CreateAgreement`+`CreateRequisition`, `CreateSession` instead of `GetRequisition`
- [x] 6.2 Update `Initiate` method — query ASPSP metadata for `MaxConsentValidity`, compute `validUntil = min(maxConsent, 180 days)`, call `StartAuth`, store `authorization_id` as `session_id`, set `auth_expires_at` from computed validity
- [x] 6.3 Update `Complete` method — accept `authorizationCode` parameter, call `CreateSession(code)`, update connection's `session_id` to new session ID, populate bank accounts with `uid` as `external_id` and `account_id` as IBAN
- [x] 6.4 Update `Reconnect` method — use `StartAuth` instead of `CreateAgreement`+`CreateRequisition`, compute consent duration dynamically
- [x] 6.5 Update `service/bank_sync_service.go` — replace GoCardless transaction field references with Enable Banking fields
- [x] 6.6 Implement composite fingerprint in `mapTransaction` — `sha256(transaction_id|booking_date|amount|currency)` when transaction_id present, `sha256(booking_date|amount|currency|first_remittance_info)` when absent
- [x] 6.7 Update `mapTransaction` description extraction — handle `RemittanceInformation` as `[]string` (use first element)

## 7. Handler Layer

- [x] 7.1 Update `handler/handler_bank.go` `CompleteConnection` — accept `code` from request body or query param, pass to service `Complete`
- [x] 7.2 Update `handler/handler_bank.go` `ListInstitutions` — map `ASPSP` response to existing `Institution` JSON shape (preserve frontend contract)

## 8. Composition Root

- [x] 8.1 Update `cmd/api/main.go` — replace `banksync.NewHTTPClient(baseURL, secretID, secretKey)` with `banksync.NewHTTPClient(baseURL, appID, keyPath)`
- [x] 8.2 Update config references from `cfg.GoCardless` to `cfg.EnableBanking`

## 9. Frontend (minor)

- [x] 9.1 Update `BankCallbackPage.tsx` — extract `code` query param from Enable Banking redirect (instead of `connection_id`) and pass to complete endpoint
- [x] 9.2 Update `ConnectBankDialog.tsx` initiate request — pass `institution_id` as ASPSP name (field semantics change, but JSON key stays same)

## 10. Cleanup & Verification

- [x] 10.1 Remove all GoCardless-specific comments and references from code
- [x] 10.2 Update `.env.example` — replace `GOCARDLESS_*` vars with `ENABLEBANKING_*` vars
- [x] 10.3 Verify backend compiles: `go build ./cmd/api`
- [x] 10.4 Verify frontend compiles: `npx tsc --noEmit`
