## Why

When users connect multiple bank accounts, money transfers between their own accounts (e.g. Revolut → main bank) are recorded as an expense on the source and income on the destination. This double-counts both spending and income, making summaries inaccurate. Transfers between own accounts should be automatically detected and excluded from financial summaries.

## What Changes

- Add `transfer` as a new `TransactionType` value alongside `expense` and `income`
- During bank sync, auto-detect internal transfers by checking if the counterparty IBAN matches any of the user's linked bank account IBANs — if so, mark the transaction as `transfer` instead of `expense`/`income`
- Exclude `transfer` transactions from spending summary aggregations (total_spent, by_category, by_participant)
- Add a "Transfers" option to the transaction type filter on the frontend so users can view/hide transfers
- The Apply Rules feature should skip `transfer` transactions (no need to categorize internal moves)

## Capabilities

### New Capabilities
- `transfer-detection`: Auto-detection of internal transfers during bank sync by matching counterparty IBAN against the user's own bank account IBANs

### Modified Capabilities
- `transaction-ledger`: Add `transfer` to the `TransactionType` enum and `transaction_type` DB enum; update type validation to accept the new value
- `spending-summary`: Exclude transactions with `type = 'transfer'` from all summary aggregations
- `bank-import`: During bank sync, resolve transaction type by checking counterparty IBAN against the user's linked accounts

## Impact

- **DB migration**: `ALTER TYPE transaction_type ADD VALUE 'transfer'`
- **Model**: `model.TypeTransfer` constant added
- **Bank sync service**: Needs access to user's bank account IBANs to compare against transaction counterparty data
- **Summary SQL queries**: Add `WHERE transaction_type != 'transfer'` filter
- **Frontend**: Type filter gains a "Transfers" option; transfer transactions shown with a distinct visual indicator
- **Apply Rules**: Skip entries on `transfer` transactions
