# Amount Range Filter

## Requirements

### API
- `GET /workspaces/{id}/transactions` accepts optional `amount_min` and `amount_max` query parameters
- Both are decimal strings (e.g. `"100"`, `"49.99"`)
- Negative values are rejected with 400
- When `amount_min` is set: only transactions where `ABS(total_amount) >= amount_min` are returned
- When `amount_max` is set: only transactions where `ABS(total_amount) <= amount_max` are returned
- When both are set: acts as a range filter (AND)
- When neither is set: no amount filtering (current behavior)

### UI
- Two number input fields ("Min" / "Max") in the transactions filter bar
- Filter is applied on input blur or Enter key press
- Changing either field resets pagination to page 0
- Empty inputs mean "no limit" on that end
