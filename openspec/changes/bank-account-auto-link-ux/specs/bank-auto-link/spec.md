## ADDED Requirements

### Requirement: Auto-link dialog after bank connection
The system SHALL display a dialog immediately after a bank OAuth connection completes, listing all discovered bank accounts with checkboxes (all pre-checked by default). The dialog SHALL allow the user to link selected accounts to the current workspace in one action.

#### Scenario: User connects a bank and sees auto-link dialog
- **WHEN** a user completes the bank OAuth flow and returns to the settings page
- **THEN** a dialog appears listing all discovered accounts with checkboxes, all checked by default, with a "Link to [Workspace Name]" button and a "Skip" option

#### Scenario: User confirms auto-link
- **WHEN** the user clicks "Link to [Workspace Name]" with 3 accounts checked
- **THEN** all 3 selected accounts are linked to the current workspace and the dialog closes with a success toast

#### Scenario: User unchecks some accounts before confirming
- **WHEN** the user unchecks 1 of 3 accounts and clicks "Link to [Workspace Name]"
- **THEN** only the 2 checked accounts are linked to the workspace

#### Scenario: User skips auto-link
- **WHEN** the user clicks "Skip" on the auto-link dialog
- **THEN** no accounts are linked and the dialog closes; accounts remain available in the "Available Accounts to Link" section

#### Scenario: Router state lost on refresh
- **WHEN** the user refreshes the settings page before interacting with the auto-link dialog
- **THEN** the dialog does not appear; accounts are still available in the manual "Available Accounts to Link" section

### Requirement: Discovered accounts passed via router state
The bank callback page SHALL pass the list of discovered accounts from the connection completion response to the settings page via React Router navigation state, so the auto-link dialog can display them without an additional API call.

#### Scenario: Successful connection passes accounts
- **WHEN** the bank OAuth callback completes successfully and receives a list of accounts
- **THEN** the callback page navigates to `/settings` with the accounts array in `location.state.newAccounts`

#### Scenario: Failed connection does not pass accounts
- **WHEN** the bank OAuth callback fails
- **THEN** the callback page navigates to `/settings` without `newAccounts` in state

### Requirement: Settings page separates connections from linked accounts
The workspace settings page SHALL display bank connections and linked accounts in two separate cards to clearly distinguish user-level bank connections from workspace-scoped linked accounts.

#### Scenario: Two distinct cards are visible
- **WHEN** a user visits the workspace settings page with at least one bank connection
- **THEN** they see a "Bank Connections" card (with connect/reconnect/delete actions) and a separate "Linked Accounts" card (with link/unlink actions)

#### Scenario: Linked accounts card shows empty state
- **WHEN** a user has bank connections but no accounts linked to the current workspace
- **THEN** the "Linked Accounts" card shows guidance text explaining how to link accounts, along with the "Available Accounts to Link" section if unlinked accounts exist

### Requirement: Linked accounts card shows available accounts to link
The "Linked Accounts" card SHALL always show an "Available Accounts to Link" sub-section when there are user-owned bank accounts not yet linked to the current workspace, regardless of whether any accounts are already linked.

#### Scenario: Unlinked accounts visible alongside linked accounts
- **WHEN** a user has 3 bank accounts total, 1 linked to the workspace
- **THEN** the "Linked Accounts" card shows the 1 linked account and an "Available Accounts to Link" section with the other 2
