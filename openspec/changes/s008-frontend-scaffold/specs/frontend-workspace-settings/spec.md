## ADDED Requirements

### Requirement: Workspace settings page
The workspace settings page SHALL allow the workspace owner to edit workspace details and manage members.

#### Scenario: Edit workspace name and description
- **WHEN** the owner edits the workspace name or description and saves
- **THEN** the app SHALL PUT to the update workspace endpoint
- **AND** the header workspace selector SHALL reflect the updated name

#### Scenario: Non-owner sees read-only settings
- **WHEN** a non-owner member visits the settings page
- **THEN** workspace details SHALL be displayed as read-only
- **AND** member management actions (add/remove) SHALL be hidden

### Requirement: Member management
The workspace settings page SHALL display the member list and allow the owner to add or remove members.

#### Scenario: List members
- **WHEN** the settings page loads
- **THEN** it SHALL display all workspace members with their name, email, role, and avatar

#### Scenario: Add member
- **WHEN** the owner enters a user ID and role and submits
- **THEN** the app SHALL POST to the add member endpoint and refresh the member list

#### Scenario: Remove member
- **WHEN** the owner clicks remove on a member
- **THEN** a confirmation dialog SHALL appear
- **AND** on confirm, the app SHALL call the remove member endpoint and refresh the list

#### Scenario: Cannot remove self
- **WHEN** the owner attempts to remove themselves
- **THEN** the remove button SHALL be disabled or hidden

### Requirement: Delete workspace
The workspace settings page SHALL allow the owner to delete the workspace.

#### Scenario: Delete workspace with confirmation
- **WHEN** the owner clicks "Delete Workspace"
- **THEN** a confirmation dialog SHALL appear with a warning about permanent data loss
- **AND** on confirm, the app SHALL call the delete endpoint and redirect to workspace selection
