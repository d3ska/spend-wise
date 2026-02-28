## Why

SpendWise is a collaborative household ledger — users share workspaces and split expenses. The current transaction form assigns 100% of the amount to the current user with no way to split costs among workspace members. Adding participant-based splits is essential for the core value proposition of shared expense tracking.

## What Changes

- Add a **split mode selector** to the create transaction dialog with three modes: "All on me" (default), "Split equally", "Custom"
- **"All on me"**: current behavior — single entry, full amount on the current user (no extra UI shown)
- **"Split equally"**: divide total equally among all workspace members, show read-only breakdown with rounding remainder on the first participant
- **"Custom"**: editable amount per member, with validation that splits sum to the total amount
- Fetch **workspace members** to populate participant list
- Rename **"entries"** to **"splits"** in UI labels for clarity

## Capabilities

### New Capabilities

- `transaction-splits`: Participant-based cost splitting in the create transaction dialog, with three split modes (all-on-me, equal, custom)

### Modified Capabilities

_(none — backend Entry model and API stay unchanged; the frontend maps UI splits to entries)_

## Impact

- **Frontend only**: `src/components/transactions/CreateTransactionDialog.tsx`
- **API hooks used**: `useWorkspace(id)` for member list, `useAuth` for current user
- **No backend changes**: the frontend maps split participants into the existing `entries[]` payload with `category_id` and `participant_id`
