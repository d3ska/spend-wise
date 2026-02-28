## ADDED Requirements

### Requirement: Category entity with workspace scope
The `model` package SHALL define a `Category` entity with typed `CategoryID`, workspace scope, name, and icon.

#### Scenario: Category fields
- **WHEN** a `Category` struct is created
- **THEN** it SHALL have fields: `ID CategoryID`, `WorkspaceID WorkspaceID`, `Name string`, `Icon string`, `CreatedAt time.Time`, `UpdatedAt time.Time`

### Requirement: Category database migration
The database SHALL have a `categories` table with a unique constraint on (workspace_id, name).

#### Scenario: Categories table schema
- **WHEN** migration 000003 is applied
- **THEN** the `categories` table SHALL have columns: `id` (BIGSERIAL PK), `workspace_id` (BIGINT NOT NULL FK to workspaces ON DELETE CASCADE), `name` (TEXT NOT NULL), `icon` (TEXT NOT NULL DEFAULT ''), `created_at`, `updated_at`
- **AND** a unique constraint on `(workspace_id, name)` SHALL exist

#### Scenario: Migration rollback
- **WHEN** migration 000003 is rolled back
- **THEN** the `categories` table SHALL be dropped

### Requirement: Category store operations
The `store` package SHALL provide category persistence with CRUD scoped to workspace.

#### Scenario: Create category
- **WHEN** a category is created with a unique name within its workspace
- **THEN** it SHALL be persisted and returned with a generated ID

#### Scenario: Duplicate category name
- **WHEN** a category is created with a name that already exists in the same workspace
- **THEN** it SHALL return an error

#### Scenario: List categories by workspace
- **WHEN** `ListByWorkspace(ctx, workspaceID)` is called
- **THEN** it SHALL return all categories belonging to that workspace

#### Scenario: Category not found
- **WHEN** `GetByID(ctx, nonExistentID)` is called
- **THEN** it SHALL return `model.ErrCategoryNotFound`

### Requirement: Category operations require workspace membership
All category operations SHALL verify that the calling user is a member of the workspace.

#### Scenario: Create category requires editor role
- **WHEN** a viewer attempts to create a category
- **THEN** it SHALL return `model.ErrInsufficientPermission`

#### Scenario: List categories requires viewer role
- **WHEN** a workspace member lists categories
- **THEN** categories SHALL be returned regardless of role (viewer, editor, owner)

#### Scenario: Non-member cannot access categories
- **WHEN** a non-member attempts any category operation
- **THEN** it SHALL return `model.ErrNotWorkspaceMember`
