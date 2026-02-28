## Architecture

Follows the existing filter pattern already established for `type`, `currency`, `category_id`, and `bank_account_id` filters — optional nullable params through every layer.

## Amount Range Filter

### SQL Layer

Add two `sqlc.narg` conditions to `ListTransactionsByWorkspace`:
```sql
AND (sqlc.narg('amount_min')::numeric IS NULL OR ABS(t.total_amount) >= sqlc.narg('amount_min'))
AND (sqlc.narg('amount_max')::numeric IS NULL OR ABS(t.total_amount) <= sqlc.narg('amount_max'))
```

`ABS()` ensures the filter works on absolute values — expenses may be stored as negative amounts but users think in terms of "transactions over 100".

### Backend Pipeline

Same pattern as existing filters:
- Handler: parse `amount_min`/`amount_max` query params as `decimal.Decimal`, reject negative
- Service: `ListTransactionsInput` gains `AmountMin *decimal.Decimal`, `AmountMax *decimal.Decimal`
- Store: `ListByWorkspaceFilter` gains same fields, converts to `pgtype.Numeric` for sqlc

### Frontend

- Two `<Input>` fields in the filter bar labeled "Min" and "Max"
- Applied on blur or Enter — avoids excessive API calls while typing
- Passed as `amount_min`/`amount_max` string params to the API

## Amount Display Format Change

### Current: Symbol-prefix
`formatMoney()` in `frontend/src/lib/money.ts` outputs `zł123.45`, `$50.00`, `€75.00`

### New: Code-suffix
Change to `123.45 PLN`, `50.00 USD`, `75.00 EUR`

This is a simple change to the `formatMoney()` function — no separate column needed. The function is used throughout the app so the change propagates everywhere automatically.

## Decisions

- **ABS filtering**: Users think in absolute amounts ("transactions over 100"), not signed values. Using `ABS()` means the filter works the same for expenses and income.
- **No separate currency column**: Currency inline with amount is more compact and standard in banking apps. Improved by switching from symbols to codes for clarity.
- **Debounce via blur/Enter**: Prevents API spam while typing numbers. Simpler than a timer-based debounce.
