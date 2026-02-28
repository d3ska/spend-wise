## ADDED Requirements

### Requirement: Frontend invite acceptance page
The frontend SHALL have a `/invite/{code}` route that handles invite link clicks.

#### Scenario: Logged-in user views invite
- **WHEN** an authenticated user navigates to `/invite/{code}`
- **THEN** the system SHALL show the workspace name, inviter name, and assigned role
- **AND** a "Join Workspace" button SHALL be displayed

#### Scenario: Logged-in user accepts invite
- **WHEN** an authenticated user clicks "Join Workspace"
- **THEN** the system SHALL call `POST /invites/{code}/accept`
- **AND** on success, navigate to the workspace dashboard

#### Scenario: Not-logged-in user views invite
- **WHEN** an unauthenticated user navigates to `/invite/{code}`
- **THEN** the system SHALL show the workspace name, inviter name, and assigned role
- **AND** a "Sign in to join" button SHALL be displayed

#### Scenario: SSO bridge via localStorage
- **WHEN** an unauthenticated user clicks "Sign in to join"
- **THEN** the invite code SHALL be stored in localStorage under key `sw_pending_invite`
- **AND** the user SHALL be redirected to the login page

#### Scenario: Auto-accept after SSO callback
- **WHEN** the auth callback completes and `sw_pending_invite` exists in localStorage
- **THEN** the system SHALL call `POST /invites/{code}/accept`
- **AND** on success, clear `sw_pending_invite` from localStorage
- **AND** navigate to the workspace dashboard

#### Scenario: Expired or invalid invite
- **WHEN** a user navigates to `/invite/{code}` and the preview returns 410 or 404
- **THEN** the system SHALL display an appropriate error message ("Invite expired" or "Invite not found")

### Requirement: Members section in workspace settings
The Settings page SHALL display a Members section showing all workspace members with their profiles and roles.

#### Scenario: Member list display
- **WHEN** a workspace member views the Settings page
- **THEN** the Members section SHALL show each member's avatar, display name, email, and role
- **AND** the total member count SHALL be displayed in the section header

#### Scenario: Owner can change member roles
- **WHEN** the workspace owner views the Members section
- **THEN** each non-owner member SHALL have a role dropdown allowing change to "editor" or "viewer"
- **AND** changing the role SHALL call `PUT /workspaces/{id}/members/{userId}` with the new role

#### Scenario: Owner can remove members
- **WHEN** the workspace owner views the Members section
- **THEN** each non-owner member SHALL have a remove button
- **AND** clicking remove SHALL show a confirmation dialog
- **AND** confirming SHALL call `DELETE /workspaces/{id}/members/{userId}`

#### Scenario: Non-owner sees read-only member list
- **WHEN** a non-owner member views the Members section
- **THEN** the member list SHALL be displayed without role dropdowns or remove buttons

#### Scenario: Owner can generate invite link
- **WHEN** the workspace owner clicks "Invite member" in the Members section
- **THEN** a dialog SHALL appear with a role selector (editor/viewer) and a "Generate link" button
- **AND** after generation, the invite URL SHALL be displayed with a "Copy" button

### Requirement: Enhanced workspace switcher
The Header workspace dropdown SHALL show member counts and allow creating new workspaces.

#### Scenario: Member count in dropdown
- **WHEN** the workspace dropdown is opened
- **THEN** each workspace SHALL display its member count (e.g., "3 members" or a people icon with count)

#### Scenario: Create workspace option
- **WHEN** the workspace dropdown is opened
- **THEN** a "+ Create workspace" option SHALL appear at the bottom, separated from the workspace list

#### Scenario: Create workspace dialog
- **WHEN** the user clicks "+ Create workspace"
- **THEN** a dialog SHALL appear with name and description fields
- **AND** on submit, a new workspace SHALL be created and selected as active

### Requirement: Frontend workspace API cleanup
The frontend SHALL remove stale `WorkspaceType` references.

#### Scenario: Workspace type removed from types
- **WHEN** the `Workspace` interface is used
- **THEN** it SHALL NOT have a `type` field

#### Scenario: Create workspace without type
- **WHEN** `useCreateWorkspace` is called
- **THEN** it SHALL send only `{ name, description }` without a `type` field
