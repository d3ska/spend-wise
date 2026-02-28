# Transaction Ledger — Delta

## Changed Requirements

### Amount Display Format
- **Before**: Currency symbol prefix — `zł123.45`, `$50.00`, `€75.00`
- **After**: Currency code suffix — `123.45 PLN`, `50.00 USD`, `75.00 EUR`
- Applies to `formatMoney()` used throughout the app (transactions table, summaries, budgets, etc.)
- Negative amounts show sign before number: `-123.45 PLN`
