## 1. Auto-Link Dialog Component

- [x] 1.1 Create `frontend/src/components/bank/AutoLinkDialog.tsx` — dialog with account checkboxes (all pre-checked), "Link to [Workspace Name]" button, "Skip" button, loading state during bulk link
- [x] 1.2 Type the `useCompleteConnection` response in `bank-connections.ts` to return `BankAccount[]` instead of `any`

## 2. BankCallbackPage — Pass Accounts via Router State

- [x] 2.1 Update `BankCallbackPage.tsx` — on success, navigate to `/settings` with `{ state: { newAccounts: data } }` instead of plain `navigate("/settings")`

## 3. Settings Page — Receive State and Open Dialog

- [x] 3.1 Update `WorkspaceSettingsPage.tsx` — read `location.state?.newAccounts`, open `AutoLinkDialog` when present, clear state after dialog closes (via `window.history.replaceState`)

## 4. Settings Page — Split into Two Cards

- [x] 4.1 Extract the "Bank Connections" section (connection list + connect/reconnect/delete) into its own card, keeping it user-level
- [x] 4.2 Extract linked accounts + available accounts into a separate "Linked Accounts" card with workspace context
- [x] 4.3 Add empty state to "Linked Accounts" card — guidance text when no accounts are linked (e.g., "Link your bank accounts to this workspace to sync transactions automatically")
- [x] 4.4 Always show "Available Accounts to Link" section in the Linked Accounts card when unlinked accounts exist (remove the conditional hide)

## 5. Verification

- [x] 5.1 Run `npx tsc --noEmit` in frontend — compiles cleanly
- [ ] 5.2 Manual test: connect a bank, verify auto-link dialog appears with accounts pre-checked, confirm linking works
- [ ] 5.3 Manual test: verify settings page shows two separate cards with clear labels
