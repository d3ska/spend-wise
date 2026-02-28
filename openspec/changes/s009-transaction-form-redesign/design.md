## Context

The current transaction create dialog in `TransactionsPage.tsx` uses a field-array of "entries" where each entry has its own category and amount. This maps directly to the backend `Entry` model but is unintuitive for users. The typical flow is: "I spent X at Y, split it with my household."

The backend API expects `entries[]` with `{category_id, amount, participant_id, note}`. This contract stays unchanged — the frontend will map the redesigned UI into the same payload shape.

## Goals / Non-Goals

**Goals:**
- Simpler form: category is a single top-level field, not repeated per split
- Auto-suggest category by matching description against workspace rules (client-side)
- Participant-based splits with smart defaults ("All on me", "Split equally", "Custom")
- Rename "entries" → "splits" in all UI labels and component names

**Non-Goals:**
- Backend API changes (entries model stays as-is)
- Editing existing transactions (out of scope — create-only)
- Server-side rule matching during manual entry (client-side only)
- Split by percentage (only absolute amounts for now)

## Decisions

### 1. Extract form into dedicated component

**Decision:** Extract the create transaction dialog into `src/components/transactions/CreateTransactionDialog.tsx`.

**Rationale:** The form is becoming complex enough (category auto-suggest, split modes, member fetching) to warrant its own file. The `TransactionsPage.tsx` stays focused on the table/list view.

### 2. Category auto-suggest via rules

**Decision:** Load rules with `useRules(workspaceId)`, filter to enabled rules, and match the description input against each rule's `match_pattern` using case-insensitive `String.includes()`. Show the highest-priority matching rule's category as a suggestion. The user can accept or override.

**Alternatives considered:**
- Regex matching: Too complex for the user-facing rule patterns, which are simple substrings
- Debounced server call: Unnecessary — rules are already cached by TanStack Query

### 3. Split mode state machine

**Decision:** Three split modes managed by a single `splitMode` state:

| Mode | Behavior |
|------|----------|
| `me` | One entry: full amount assigned to current user. No split UI shown. |
| `equal` | Entries auto-generated: `total / memberCount` for each member. Read-only amounts. |
| `custom` | Editable amount fields per member. User adds/removes participants. |

On mode change, splits are recalculated from the current total amount.

### 4. Mapping UI splits → API entries

**Decision:** On submit, map each split to an entry:
```
entry = {
  category_id: formCategory,   // single top-level category
  participant_id: split.userId,
  amount: { amount: split.amount, currency },
  note: ""
}
```
The `total_amount` is the sum of all split amounts.

### 5. Workspace members source

**Decision:** Use the existing `useWorkspace(id)` hook which returns `{ members: WorkspaceMember[] }`. Members include `user_id` and `role`. We need display names, so we'll use the member data available from the workspace detail endpoint.

## Risks / Trade-offs

- **Rule matching is simple substring** → May produce false positives for short patterns. Acceptable for v1; users can override the suggestion.
- **Equal split rounding** → Dividing amounts may produce rounding remainders (e.g., 100 / 3). Assign remainder cents to the first participant. Use `toFixed(2)` arithmetic.
- **No percentage splits** → Keeps the UI simple. Custom mode uses absolute amounts. Can be added later if needed.
