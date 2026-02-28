## Why

The current "New Transaction" dialog treats each entry as a separate category+amount row, which is confusing for the primary use case: a single expense paid by one person that may be split among workspace members. Users expect to pick a category once, enter a total, and optionally split costs — not manage a list of ledger entries. Additionally, the workspace's categorization rules (used during bank import) are not leveraged during manual entry, missing an opportunity to auto-suggest categories as the user types a description.

## What Changes

- Move **category** from per-entry to a **top-level field** on the transaction form
- **Auto-suggest category** from loaded rules when the user types a description (client-side substring matching against `match_pattern` of enabled rules)
- Rename **"Entries"** to **"Splits"** in the UI for clarity
- Redesign splits as **participant-based**: each split assigns an amount to a workspace member
- Add a **split mode selector**: "All on me" (default), "Split equally", "Custom"
  - "All on me" — single entry, 100% of the amount on the current user
  - "Split equally" — divide total equally among all workspace members
  - "Custom" — manually assign amounts per participant
- **Fetch workspace members** to populate the split participant list

## Capabilities

### New Capabilities

- `transaction-form-ux`: Client-side transaction form redesign covering category placement, rule-based auto-suggest, and participant-based split modes

### Modified Capabilities

_(none — backend Entry model and API stay unchanged; the frontend maps UI splits to entries)_

## Impact

- **Frontend only**: `src/pages/TransactionsPage.tsx` (create dialog), possibly extracted into its own component
- **API hooks used**: `useRules`, `useWorkspace` (for members), `useCategories`, `useCreateTransaction`
- **No backend changes**: the frontend maps the new UI model (single category + participant splits) into the existing `entries[]` payload with `category_id` and `participant_id`
