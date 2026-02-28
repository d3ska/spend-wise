## ADDED Requirements

### Requirement: Workspace entity with typed ID and validation
The `model` package SHALL define a `Workspace` entity with typed `WorkspaceID`, `WorkspaceType` enum (private|shared), mandatory name and description, and a `Validate()` method.

#### Scenario: Workspace with valid fields
- **WHEN** `Validate()` is called on a Workspace with non-empty name and description and a valid type
- **THEN** it SHALL return no error

#### Scenario: Workspace missing name
- **WHEN** `Validate()` is called on a Workspace with empty name
- **THEN** it SHALL return `model.ErrWorkspaceNameRequired`

#### Scenario: Workspace missing description
- **WHEN** `Validate()` is called on a Workspace with empty description
- **THEN** it SHALL return `model.ErrWorkspaceDescriptionRequired`

#### Scenario: Workspace with invalid type
- **WHEN** `Validate()` is called on a Workspace with a type other than "private" or "shared"
- **THEN** it SHALL return `model.ErrInvalidWorkspaceType`

### Requirement: WorkspaceMember with role-based permissions
The `model` package SHALL define a `WorkspaceMember` entity with `MemberRole` enum (owner|editor|viewer). Roles SHALL be comparable by level: owner > editor > viewer.

#### Scenario: Role level comparison
- **WHEN** `RoleOwner.Level()` is compared to `RoleEditor.Level()`
- **THEN** owner SHALL have a higher level than editor
- **AND** editor SHALL have a higher level than viewer

#### Scenario: CanEdit check
- **WHEN** `CanEdit()` is called on a member with role editor
- **THEN** it SHALL return true
- **AND** `CanEdit()` on a viewer SHALL return false

### Requirement: Workspace database migration
The database SHALL have `workspaces` and `workspace_members` tables with PostgreSQL enums for type and role.

#### Scenario: Workspaces table schema
- **WHEN** migration 000002 is applied
- **THEN** the `workspaces` table SHALL have columns: `id` (BIGSERIAL PK), `name` (TEXT NOT NULL), `description` (TEXT NOT NULL), `type` (workspace_type NOT NULL), `owner_id` (BIGINT NOT NULL FK to users), `created_at`, `updated_at`

#### Scenario: Workspace members table schema
- **WHEN** migration 000002 is applied
- **THEN** the `workspace_members` table SHALL have composite PK `(workspace_id, user_id)`, `role` (member_role NOT NULL), `created_at`
- **AND** it SHALL CASCADE on workspace deletion

#### Scenario: Migration rollback
- **WHEN** migration 000002 is rolled back
- **THEN** `workspace_members`, `workspaces` tables and `workspace_type`, `member_role` enum types SHALL be dropped

### Requirement: Workspace store operations
The `store` package SHALL provide workspace persistence with CRUD, member management, and user-scoped listing.

#### Scenario: Create workspace
- **WHEN** a workspace is created
- **THEN** it SHALL be persisted and the created workspace SHALL be returned with a generated ID

#### Scenario: List workspaces by user
- **WHEN** `ListByUser(ctx, userID)` is called
- **THEN** it SHALL return all workspaces where the user is a member

#### Scenario: Get workspace not found
- **WHEN** `GetByID(ctx, nonExistentID)` is called
- **THEN** it SHALL return `model.ErrWorkspaceNotFound`

#### Scenario: Add member
- **WHEN** `AddMember(ctx, workspaceID, userID, role)` is called
- **THEN** a workspace_members row SHALL be inserted

#### Scenario: Get member
- **WHEN** `GetMember(ctx, workspaceID, userID)` is called for an existing member
- **THEN** it SHALL return the member with their role

#### Scenario: Get member not found
- **WHEN** `GetMember(ctx, workspaceID, userID)` is called for a non-member
- **THEN** it SHALL return `model.ErrNotWorkspaceMember`

### Requirement: Workspace service with permission enforcement
The `service` package SHALL provide workspace operations with role-based permission checks.

#### Scenario: Create workspace auto-adds owner
- **WHEN** `Create(ctx, userID, input)` is called
- **THEN** the workspace SHALL be created
- **AND** the user SHALL be added as a member with role owner

#### Scenario: Update requires editor role
- **WHEN** `Update(ctx, userID, wsID, input)` is called by a viewer
- **THEN** it SHALL return `model.ErrInsufficientPermission`

#### Scenario: Delete requires owner role
- **WHEN** `Delete(ctx, userID, wsID)` is called by an editor
- **THEN** it SHALL return `model.ErrInsufficientPermission`

#### Scenario: Add member requires owner role
- **WHEN** `AddMember(ctx, userID, wsID, targetUserID, role)` is called by an editor
- **THEN** it SHALL return `model.ErrInsufficientPermission`

#### Scenario: Non-member access denied
- **WHEN** any workspace operation is called by a user who is not a member
- **THEN** it SHALL return `model.ErrNotWorkspaceMember`
