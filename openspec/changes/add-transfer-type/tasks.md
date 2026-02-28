## 1. Database Migration

- [x] 1.1 Create migration `000010_add_transfer_type.up.sql` with `ALTER TYPE transaction_type ADD VALUE 'transfer'`
- [x] 1.2 Create migration `000010_add_transfer_type.down.sql` as no-op (comment explaining PostgreSQL cannot remove enum values)

## 2. Model & Constants

- [x] 2.1 Add `TypeTransfer TransactionType = "transfer"` constant in `model/transaction.go`

## 3. Enable Banking Client

- [x] 3.1 Add `Account *AccountIdentification` field to `banksync.PartyIdentification` struct in `banksync/client.go`

## 4. Bank Sync Service

- [x] 4.1 Add `buildUserIBANSet` helper in `service/bank_sync_service.go` that loads all bank accounts for a user via `acctStore` and returns a `map[string]bool` of non-nil IBANs
- [x] 4.2 Update `SyncAll` to build the IBAN set once per connection and pass it to `SyncBankAccount`
- [x] 4.3 Update `SyncBankAccount` signature to accept `ownIBANs map[string]bool`
- [x] 4.4 Update `mapTransaction` signature to accept `ownIBANs map[string]bool`
- [x] 4.5 Update `mapTransaction` type logic: after determining CRDT/DBIT, check counterparty IBAN (creditor for DBIT, debtor for CRDT) against `ownIBANs` — if match, set type to `model.TypeTransfer`

## 5. Handler Validation

- [x] 5.1 Update `ListTransactions` handler type validation in `handler/handler_transaction.go` to accept `transfer` alongside `expense` and `income`

## 6. Apply Rules

- [x] 6.1 Update `ListAllTransactionEntries` SQL query in `db/queries/transactions.sql` to also return `t.transaction_type`
- [x] 6.2 Update generated code in `store/transactions.sql.go` — add `TransactionType` to `ListAllTransactionEntriesRow` and scan
- [x] 6.3 Update `TransactionEntryRow` in `store/transaction_store.go` to include `Type model.TransactionType`
- [x] 6.4 Update `ApplyRules` in `service/transaction_service.go` to skip entries where `e.Type == model.TypeTransfer`

## 7. Frontend

- [x] 7.1 Add `"transfer"` to `TransactionType` union in `frontend/src/types/index.ts`
- [x] 7.2 Add "Transfers" option to the type filter toggle button group in `TransactionsPage.tsx`
- [x] 7.3 Show a distinct badge for transfer transactions (e.g. "Transfer" with `ArrowLeftRight` icon) in the transaction row

## 8. Verification

- [x] 8.1 Run `go build ./cmd/api` — backend compiles
- [x] 8.2 Run `go test -count=1 ./...` — all tests pass
- [x] 8.3 Run `npx tsc --noEmit` — frontend compiles
