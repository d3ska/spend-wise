## 1. Extract and Scaffold

- [x] 1.1 Create `src/components/transactions/CreateTransactionDialog.tsx` — extract dialog from `TransactionsPage.tsx` into its own component with the same props/behavior
- [x] 1.2 Update `TransactionsPage.tsx` to import and use the extracted `CreateTransactionDialog`

## 2. Form Redesign

- [x] 2.1 Move category to a top-level select field (outside of entries/splits section) using the existing `useCategories` hook
- [x] 2.2 Add split mode selector with three options: "All on me" (default), "Split equally", "Custom" — using a button group or segmented control
- [x] 2.3 Fetch workspace members via `useWorkspace(workspaceId)` and expose member list for split participants
- [x] 2.4 Implement "All on me" mode — hide split details, assign full amount to current user on submit
- [x] 2.5 Implement "Split equally" mode — show read-only member list with auto-calculated equal amounts (handle rounding remainder)
- [x] 2.6 Implement "Custom" mode — show editable amount fields per member, validate that splits sum to total
- [x] 2.7 Rename all "entry/entries" labels to "split/splits" in the dialog UI

## 3. Category Auto-Suggest

- [x] 3.1 Load workspace rules via `useRules(workspaceId)` in the dialog
- [x] 3.2 On description input change, match against enabled rules' `match_pattern` (case-insensitive substring), auto-select the highest-priority match's `target_category_id`
- [x] 3.3 Allow user to override the auto-suggested category by manually selecting a different one

## 4. Submit Mapping

- [x] 4.1 Map form data to API payload: single `category_id` applied to all entries, each split becomes an entry with `participant_id` and `amount`
- [x] 4.2 Verify form validation (Zod schema update): total amount required, category required, splits must sum correctly in custom mode

## 5. Polish

- [x] 5.1 Ensure form resets correctly when dialog closes and reopens
- [x] 5.2 Run `npm run build` and `npm run lint` to verify clean build
