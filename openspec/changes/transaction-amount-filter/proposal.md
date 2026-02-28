## Why

The transactions list currently lacks the ability to filter by amount range, making it hard to find specific transactions (e.g. "show me all transactions over 500 PLN"). Additionally, the amount display uses currency symbols (e.g. `zł123.45`) which can be unclear — not everyone recognizes `zł` as PLN. Switching to `123.45 PLN` format is more universally readable.

## What Changes

- **Amount range filter**: Add optional `amount_min` and `amount_max` query parameters to the list transactions API, filtering on `ABS(total_amount)` so it works intuitively regardless of sign
- **Improved amount format**: Change money display from symbol-prefix (`zł123.45`) to code-suffix (`123.45 PLN`) across the transactions table for clarity

## Capabilities

### New Capabilities

- `amount-range-filter`: Server-side filtering of transactions by minimum and/or maximum absolute amount, with corresponding UI inputs

### Modified Capabilities

- `transaction-ledger`: Amount display format changes from symbol-prefix to code-suffix (e.g. `zł123.45` → `123.45 PLN`)

## Impact

- **Backend**: SQL query, sqlc generated code, store/service/handler layers gain two new optional filter params
- **Frontend**: Transactions page gains min/max amount inputs; `formatMoney()` updated for new format; API client gets new params
- **API**: `GET /workspaces/{id}/transactions` gains `amount_min` and `amount_max` query parameters
