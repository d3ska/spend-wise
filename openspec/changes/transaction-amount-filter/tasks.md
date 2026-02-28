## 1. Backend — Amount Range Filter

- [ ] 1.1 Update `backend/db/queries/transactions.sql` — add `amount_min` and `amount_max` `sqlc.narg` conditions with `ABS()` to `ListTransactionsByWorkspace`
- [ ] 1.2 Run `make sqlc` to regenerate `backend/store/transactions.sql.go`
- [ ] 1.3 Update `backend/store/transaction_store.go` — add `AmountMin *decimal.Decimal` and `AmountMax *decimal.Decimal` to `ListByWorkspaceFilter`, convert to `pgtype.Numeric` in `ListByWorkspace()`
- [ ] 1.4 Update `backend/service/transaction_service.go` — add `AmountMin *decimal.Decimal` and `AmountMax *decimal.Decimal` to `ListTransactionsInput`, pass through to store filter
- [ ] 1.5 Update `backend/handler/handler_transaction.go` — parse optional `amount_min` and `amount_max` query params in `ListTransactions()`, validate non-negative

## 2. Frontend — Amount Range Filter

- [ ] 2.1 Update `frontend/src/api/transactions.ts` — add `amount_min?: string` and `amount_max?: string` to `TransactionListParams`
- [ ] 2.2 Update `frontend/src/pages/TransactionsPage.tsx` — add `amountMin`/`amountMax` state, two `<Input>` fields in filter bar, apply on blur/Enter, pass to API, reset page on change

## 3. Frontend — Amount Display Format

- [ ] 3.1 Update `frontend/src/lib/money.ts` — change `formatMoney()` from symbol-prefix (`zł123.45`) to code-suffix (`123.45 PLN`)

## 4. Verification

- [ ] 4.1 Run `make build` to verify backend compilation
- [ ] 4.2 Run `make test` to verify all tests pass
- [ ] 4.3 Manual test: verify amount displays as `123.45 PLN` format across the app
- [ ] 4.4 Manual test: filter by min amount only, max amount only, and both
- [ ] 4.5 Manual test: verify empty inputs show all transactions (no filtering)
