## ADDED Requirements

### Requirement: Workspace invites database table
The database SHALL have a `workspace_invites` table for storing invite links.

#### Scenario: Invites table schema
- **WHEN** the invite migration is applied
- **THEN** the `workspace_invites` table SHALL have columns: `id` (BIGSERIAL PK), `workspace_id` (BIGINT NOT NULL FK to workspaces ON DELETE CASCADE), `code` (TEXT NOT NULL UNIQUE), `role` (member_role NOT NULL), `created_by` (BIGINT NOT NULL FK to users), `expires_at` (TIMESTAMPTZ NOT NULL), `used_by` (BIGINT FK to users, nullable), `used_at` (TIMESTAMPTZ, nullable), `created_at` (TIMESTAMPTZ NOT NULL DEFAULT NOW())

#### Scenario: Migration rollback
- **WHEN** the invite migration is rolled back
- **THEN** the `workspace_invites` table SHALL be dropped

### Requirement: Invite code generation
The system SHALL generate cryptographically secure, URL-safe invite codes using 32 bytes of `crypto/rand` encoded as base64url (no padding).

#### Scenario: Code uniqueness
- **WHEN** an invite is created
- **THEN** the code SHALL be unique across all invites

#### Scenario: Code format
- **WHEN** an invite code is generated
- **THEN** it SHALL be URL-safe (base64url, no padding characters)

### Requirement: Create invite endpoint
The system SHALL provide `POST /api/v1/workspaces/{id}/invites` to create an invite link. Only workspace owners SHALL be able to create invites.

#### Scenario: Owner creates invite with role
- **WHEN** a workspace owner sends `POST /workspaces/{id}/invites` with `{ "role": "editor" }`
- **THEN** the system SHALL create an invite with the specified role and return `201` with `{ "code", "role", "expires_at", "invite_url" }`

#### Scenario: Owner creates invite with default expiry
- **WHEN** a workspace owner creates an invite without specifying `expires_in_hours`
- **THEN** the invite SHALL expire in 7 days (168 hours) by default

#### Scenario: Owner creates invite with custom expiry
- **WHEN** a workspace owner creates an invite with `{ "role": "viewer", "expires_in_hours": 24 }`
- **THEN** the invite SHALL expire in 24 hours from creation

#### Scenario: Non-owner cannot create invite
- **WHEN** a workspace editor or viewer sends `POST /workspaces/{id}/invites`
- **THEN** the system SHALL return `403 Forbidden`

#### Scenario: Invalid role rejected
- **WHEN** an invite is created with role "owner"
- **THEN** the system SHALL return `400 Bad Request` (only editor and viewer are allowed)

### Requirement: Preview invite endpoint (public)
The system SHALL provide `GET /api/v1/invites/{code}` as a public endpoint (no auth required) to preview an invite before accepting.

#### Scenario: Valid invite preview
- **WHEN** a valid, unexpired, unused invite code is accessed
- **THEN** the system SHALL return `200` with `{ "workspace_name", "inviter_name", "role", "expires_at" }`

#### Scenario: Expired invite preview
- **WHEN** an expired invite code is accessed
- **THEN** the system SHALL return `410 Gone` with `{ "error": "invite has expired" }`

#### Scenario: Used invite preview
- **WHEN** a used (single-use consumed) invite code is accessed
- **THEN** the system SHALL return `410 Gone` with `{ "error": "invite has already been used" }`

#### Scenario: Unknown invite code
- **WHEN** a non-existent invite code is accessed
- **THEN** the system SHALL return `404 Not Found`

### Requirement: Accept invite endpoint
The system SHALL provide `POST /api/v1/invites/{code}/accept` to accept an invite and join the workspace. Authentication is required.

#### Scenario: Accept valid invite
- **WHEN** an authenticated user accepts a valid, unexpired, unused invite
- **THEN** the user SHALL be added to the workspace with the invite's role
- **AND** the invite SHALL be marked as used (`used_by`, `used_at` set)
- **AND** the system SHALL return `200` with the workspace details

#### Scenario: Already a member
- **WHEN** an authenticated user accepts an invite for a workspace they already belong to
- **THEN** the system SHALL return `409 Conflict` with `{ "error": "already a member of this workspace" }`

#### Scenario: Accept expired invite
- **WHEN** an authenticated user tries to accept an expired invite
- **THEN** the system SHALL return `410 Gone`

#### Scenario: Accept used invite
- **WHEN** an authenticated user tries to accept an already-used invite
- **THEN** the system SHALL return `410 Gone`

#### Scenario: Accept non-existent invite
- **WHEN** an authenticated user tries to accept a non-existent code
- **THEN** the system SHALL return `404 Not Found`

### Requirement: Invite service validation
The invite service SHALL validate all business rules before creating or accepting invites.

#### Scenario: Workspace must exist
- **WHEN** an invite is created for a non-existent workspace
- **THEN** the system SHALL return `404 Not Found`

#### Scenario: Invite creator must be owner
- **WHEN** a non-owner tries to create an invite
- **THEN** the system SHALL return `model.ErrInsufficientPermission`

### Requirement: Invite store operations
The `store` package SHALL provide invite persistence operations.

#### Scenario: Insert invite
- **WHEN** `InsertInvite` is called with valid params
- **THEN** the invite SHALL be persisted and returned with generated ID

#### Scenario: Get invite by code
- **WHEN** `GetInviteByCode` is called with an existing code
- **THEN** it SHALL return the invite with workspace name and inviter display name

#### Scenario: Get invite by code not found
- **WHEN** `GetInviteByCode` is called with a non-existent code
- **THEN** it SHALL return `model.ErrInviteNotFound`

#### Scenario: Mark invite as used
- **WHEN** `MarkInviteUsed` is called with invite ID and user ID
- **THEN** the invite's `used_by` and `used_at` SHALL be set
