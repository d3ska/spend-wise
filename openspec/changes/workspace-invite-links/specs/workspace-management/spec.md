## MODIFIED Requirements

### Requirement: Workspace entity with typed ID and validation
The `model` package SHALL define a `Workspace` entity with typed `WorkspaceID`, mandatory name and description, and a `Validate()` method.

#### Scenario: Workspace with valid fields
- **WHEN** `Validate()` is called on a Workspace with non-empty name and description
- **THEN** it SHALL return no error

#### Scenario: Workspace missing name
- **WHEN** `Validate()` is called on a Workspace with empty name
- **THEN** it SHALL return `model.ErrWorkspaceNameRequired`

#### Scenario: Workspace missing description
- **WHEN** `Validate()` is called on a Workspace with empty description
- **THEN** it SHALL return `model.ErrWorkspaceDescriptionRequired`

### Requirement: Workspace database migration
The database SHALL have `workspaces` and `workspace_members` tables with a PostgreSQL enum for role.

#### Scenario: Workspaces table schema
- **WHEN** migration 000002 is applied
- **THEN** the `workspaces` table SHALL have columns: `id` (BIGSERIAL PK), `name` (TEXT NOT NULL), `description` (TEXT NOT NULL), `owner_id` (BIGINT NOT NULL FK to users), `created_at`, `updated_at`

#### Scenario: Workspace members table schema
- **WHEN** migration 000002 is applied
- **THEN** the `workspace_members` table SHALL have composite PK `(workspace_id, user_id)`, `role` (member_role NOT NULL), `created_at`
- **AND** it SHALL CASCADE on workspace deletion

#### Scenario: Migration rollback
- **WHEN** migration 000002 is rolled back
- **THEN** `workspace_members`, `workspaces` tables and `member_role` enum type SHALL be dropped

## ADDED Requirements

### Requirement: Enriched member listing with user profiles
The system SHALL provide a member listing that includes user profile information (display name, email, avatar URL) alongside membership data.

#### Scenario: List members with profiles
- **WHEN** `GET /workspaces/{id}/members` is called by a workspace member (viewer+)
- **THEN** it SHALL return an array of members, each with `user_id`, `role`, `display_name`, `email`, `avatar_url`, `created_at`

#### Scenario: List members requires membership
- **WHEN** `GET /workspaces/{id}/members` is called by a non-member
- **THEN** it SHALL return `403 Forbidden`

### Requirement: Update member role
The system SHALL allow workspace owners to change a member's role.

#### Scenario: Owner updates member role
- **WHEN** a workspace owner sends `PUT /workspaces/{id}/members/{userId}` with `{ "role": "editor" }`
- **THEN** the member's role SHALL be updated
- **AND** the system SHALL return `200` with the updated member details

#### Scenario: Cannot update own role
- **WHEN** a workspace owner tries to update their own role
- **THEN** the system SHALL return `400 Bad Request`

#### Scenario: Cannot set role to owner
- **WHEN** a workspace owner tries to set another member's role to "owner"
- **THEN** the system SHALL return `400 Bad Request`

#### Scenario: Non-owner cannot update roles
- **WHEN** a non-owner member sends `PUT /workspaces/{id}/members/{userId}`
- **THEN** the system SHALL return `403 Forbidden`

### Requirement: Member count in workspace listing
The workspace listing endpoint SHALL include a member count for each workspace.

#### Scenario: Workspace list includes member count
- **WHEN** `GET /workspaces` is called
- **THEN** each workspace in the response SHALL include a `member_count` integer field

#### Scenario: Single member workspace
- **WHEN** a workspace has only the owner
- **THEN** `member_count` SHALL be `1`

## REMOVED Requirements

### Requirement: Workspace with invalid type
**Reason**: `WorkspaceType` (private/shared) has been removed. The membership system determines sharing implicitly — a workspace with one member is personal, a workspace with multiple members is shared.
**Migration**: No migration needed. Field already removed from backend code, migration, and API. Frontend cleanup included in this change.
