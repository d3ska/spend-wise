## Why

After connecting a bank via OAuth, the user is redirected back to the settings page with no indication that they still need to manually link individual bank accounts to their workspace. The "Available Accounts to Link" section is easy to miss, and the entire linking concept is unclear. Users expect that connecting a bank to a workspace means those accounts feed into that workspace automatically.

## What Changes

- Show an **auto-link dialog** immediately after a bank connection completes, listing discovered accounts with checkboxes (all pre-checked) so the user can link them to the current workspace in one click
- The `BankCallbackPage` already receives the accounts list from the backend response — pass it to the settings page via navigation state instead of discarding it
- Split the "Bank Connections" card into two visually distinct sections: a user-level "Your Bank Connections" card (connect/reconnect/delete) and a workspace-scoped "Linked Accounts" card (link/unlink), so the ownership model is clear
- Add an empty state to the "Linked Accounts" card with a prompt guiding the user to link accounts

## Capabilities

### New Capabilities
- `bank-auto-link`: Auto-link dialog shown after bank OAuth completion, and settings page restructuring for clarity

### Modified Capabilities

(none — no existing spec-level requirements change; the bank-import spec covers sync behavior which is unaffected)

## Impact

- **Frontend only** — no backend changes needed (the `POST /bank-connections/complete` endpoint already returns discovered accounts)
- `BankCallbackPage.tsx` — pass accounts via router state
- `WorkspaceSettingsPage.tsx` — split into two cards, receive router state, open auto-link dialog
- New `AutoLinkDialog.tsx` component
- No API changes, no database changes
