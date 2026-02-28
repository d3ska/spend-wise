## Context

The create transaction dialog (`CreateTransactionDialog.tsx`) currently creates a single entry assigned to the current user. The backend already supports multiple entries per transaction with `participant_id` fields — no API changes are needed. This is a frontend-only feature to expose participant-based splitting in the UI.

The current form fields are: description, date, amount, category (with rule auto-suggest), notes. The splits feature adds a split mode selector below notes.

## Goals / Non-Goals

**Goals:**
- Allow users to split a transaction among workspace members
- Three split modes: "All on me" (default), "Split equally", "Custom"
- Keep the default experience simple — "All on me" shows no extra UI

**Non-Goals:**
- Percentage-based splits (only absolute amounts)
- Per-split categories (single category applies to all entries)
- Backend API changes
- Editing splits on existing transactions

## Decisions

### 1. Split mode as local component state

**Decision:** Manage `splitMode` as a `useState` in the dialog, not as part of the Zod form schema.

**Rationale:** Split mode is a UI control that determines how entries are generated on submit — it's not a form field that gets validated or submitted. Keeping it outside the form avoids unnecessary schema complexity.

### 2. Workspace members from existing hook

**Decision:** Use `useWorkspace(workspaceId)` which returns `{ members: WorkspaceMember[] }`. Each member has `user_id` and `role`.

**Alternative considered:** A dedicated `/members` endpoint — unnecessary since workspace detail already includes members.

**Limitation:** `WorkspaceMember` has no `display_name`. For the current user we show `"Name (you)"` from `useAuth()`. For others we show `"Member #id"`. This is acceptable for now; a future change could enrich member data.

### 3. Equal split rounding

**Decision:** Use floor-based rounding: `Math.floor(total / count * 100) / 100` per person, assign the remainder to the first participant.

**Example:** 100.00 / 3 = 33.33, 33.33, 33.34 (first person gets the extra cent).

### 4. Custom split validation

**Decision:** Validate that custom split amounts sum to the total (within 0.01 tolerance). Show an inline error and disable the submit button if they don't match. This is UI-only validation — the backend doesn't enforce split sums.

### 5. Entry mapping on submit

**Decision:** On submit, map each split to an API entry:
```
{ category_id, participant_id: split.userId, amount: { amount, currency } }
```
For "All on me" mode, a single entry with the current user's ID (same as current behavior).

## Risks / Trade-offs

- **Member display names** → Only the current user has a name; others show "Member #id". Acceptable for v1 since most shared workspaces are small households where users know each other.
- **No percentage splits** → Keeps UI simple. Users must manually calculate for percentage-based splits. Can be added later.
- **Rounding precision** → JavaScript float math with `toFixed(2)` may have edge cases. Using floor + remainder approach mitigates this for common cases.
