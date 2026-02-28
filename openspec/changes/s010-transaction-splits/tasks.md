## 1. State and Data

- [x] 1.1 Add `splitMode` state (`"me" | "equal" | "custom"`) defaulting to `"me"` in `CreateTransactionDialog`
- [x] 1.2 Add `customAmounts` state (`Map<number, string>`) for custom split values
- [x] 1.3 Fetch workspace members via `useWorkspace(workspaceId)` and memoize the members array
- [x] 1.4 Create `memberName(userId)` helper — current user shows `"Name (you)"`, others show `"Member #id"`

## 2. Split Mode UI

- [x] 2.1 Add "Splits" label and button group with three options: "All on me", "Split equally", "Custom"
- [x] 2.2 Highlight the active mode button with `variant="default"`, inactive with `variant="outline"`
- [x] 2.3 Hide split detail UI when mode is "All on me"

## 3. Equal Split Mode

- [x] 3.1 Compute equal splits with `Math.floor(total / count * 100) / 100`, assign remainder to first participant
- [x] 3.2 Show read-only member list with name and calculated amount when mode is "Split equally"
- [x] 3.3 Recalculate when total amount changes

## 4. Custom Split Mode

- [x] 4.1 Show editable amount input per member when mode is "Custom"
- [x] 4.2 Pre-fill custom amounts with equal split values when switching to custom mode
- [x] 4.3 Validate that custom amounts sum to total (within 0.01 tolerance) — show inline error and disable submit on mismatch

## 5. Submit Mapping

- [x] 5.1 "All on me" — single entry with current user's `participant_id` and full amount (existing behavior)
- [x] 5.2 "Split equally" — one entry per member with calculated amount
- [x] 5.3 "Custom" — one entry per member with non-zero amount
- [x] 5.4 All entries use the single top-level `category_id`

## 6. Reset and Polish

- [x] 6.1 Reset `splitMode` to "me" and clear `customAmounts` when dialog closes
- [x] 6.2 Run `npm run build` and `npm run lint` to verify clean build
