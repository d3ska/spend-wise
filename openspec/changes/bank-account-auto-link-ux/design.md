## Context

SpendWise users connect bank accounts via OAuth (Enable Banking). Bank connections are user-level (not workspace-scoped), while the transactions they produce are workspace-scoped through an explicit many-to-many `workspace_bank_accounts` join table. After the OAuth flow, the user must manually link each discovered account to a workspace — a step that's easy to miss.

The backend `POST /bank-connections/complete` already returns the list of discovered `BankAccount` objects. The frontend currently discards this response.

## Goals / Non-Goals

**Goals:**
- Make linking bank accounts to the current workspace a natural, guided step after connecting a bank
- Clearly separate user-level concerns (bank connections) from workspace-level concerns (linked accounts) in the settings UI
- Zero backend changes

**Non-Goals:**
- Changing the bank connection or sync data model
- Auto-linking without user confirmation (user should see which accounts are being linked)
- Historical backfill when linking to a new workspace (separate concern)

## Decisions

### 1. Pass discovered accounts via React Router state

**Decision:** `BankCallbackPage` passes the accounts array from the `CompleteByCode` response as `navigate("/settings", { state: { newAccounts: accounts } })`. `WorkspaceSettingsPage` reads `location.state?.newAccounts` and opens the auto-link dialog if present.

**Rationale:** No new API calls needed. Router state is ephemeral — it disappears on manual navigation to `/settings`, so the dialog only appears when coming from the callback. No localStorage coordination needed.

**Alternatives considered:**
- *Query parameter flag + refetch*: Would require an extra API call and couldn't carry the accounts list
- *localStorage*: Would need cleanup logic; router state is simpler and auto-cleans

### 2. Auto-link dialog with pre-checked checkboxes

**Decision:** Show a dialog with all discovered accounts listed as checkboxes, all checked by default. User clicks "Link to [Workspace Name]" to link selected accounts in bulk. A "Skip" button closes without linking.

**Rationale:** Matches the Slack/Notion pattern of "here's what we found, confirm to proceed." Pre-checking all accounts is the right default since users usually want all accounts linked. Checkboxes give control without extra steps.

### 3. Split settings into two cards

**Decision:** Replace the single "Bank Connections" card with:
1. **"Bank Connections"** card — user-level: list connections, connect/reconnect/delete buttons
2. **"Linked Accounts"** card — workspace-level: list linked accounts with unlink buttons, show "Available Accounts to Link" sub-section, empty state with guidance text

**Rationale:** The current single-card layout mixes user-level and workspace-level concepts. Separating them makes the ownership model visible and the link/unlink actions discoverable.

### 4. Bulk link via sequential mutation calls

**Decision:** When the user confirms the auto-link dialog, call `POST /workspaces/{id}/bank-accounts` sequentially for each selected account (reusing the existing `useLinkBankAccount` hook). Show a loading state during the operation.

**Rationale:** The existing link endpoint handles one account at a time. A bulk endpoint would require backend changes (non-goal). At household scale (2-5 accounts per connection), sequential calls complete in under a second.

## Risks / Trade-offs

- **[Router state lost on refresh]** If the user refreshes `/settings` before interacting with the dialog, the `newAccounts` state is lost. → Acceptable: the accounts still appear in "Available Accounts to Link" and can be linked manually. The dialog is a convenience, not the only path.
- **[Reconnect flow doesn't show dialog]** Reconnecting an existing connection may discover new accounts but uses a different code path. → For now, only the initial connection shows the auto-link dialog. Reconnect already preserves existing links via IBAN-based account reuse.
