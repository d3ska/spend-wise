## ADDED Requirements

### Requirement: CategorizationRule entity with tiered scoping
The `model` package SHALL define a `CategorizationRule` entity with typed `RuleID`, `RuleScope` enum (system|user|workspace), nullable OwnerID (for user scope), nullable WorkspaceID (for workspace scope), MatchPattern string, TargetCategoryID, Priority int, and Enabled boolean (default true).

#### Scenario: Rule fields
- **WHEN** a `CategorizationRule` struct is created
- **THEN** it SHALL have fields: `ID RuleID`, `Scope RuleScope`, `OwnerID *UserID`, `WorkspaceID *WorkspaceID`, `MatchPattern string`, `TargetCategoryID CategoryID`, `Priority int`, `Enabled bool`, `CreatedAt time.Time`, `UpdatedAt time.Time`

#### Scenario: Scope precedence
- **WHEN** `ScopePrecedence()` is called on rules with different scopes
- **THEN** workspace scope SHALL return 3, user scope SHALL return 2, and system scope SHALL return 1

### Requirement: Rule database migration
The database SHALL have a `categorization_rules` table with a `rule_scope` PostgreSQL enum and CHECK constraints enforcing scope-FK consistency.

#### Scenario: Rules table schema
- **WHEN** migration 000006 is applied
- **THEN** the `categorization_rules` table SHALL have columns: `id` (BIGSERIAL PK), `scope` (rule_scope NOT NULL), `owner_id` (BIGINT FK to users), `workspace_id` (BIGINT FK to workspaces ON DELETE CASCADE), `match_pattern` (TEXT NOT NULL), `target_category_id` (BIGINT NOT NULL FK to categories), `priority` (INT NOT NULL DEFAULT 0), `enabled` (BOOLEAN NOT NULL DEFAULT TRUE), `created_at`, `updated_at`
- **AND** a CHECK constraint SHALL enforce: workspace scope requires workspace_id NOT NULL, user scope requires owner_id NOT NULL

#### Scenario: Migration rollback
- **WHEN** migration 000006 is rolled back
- **THEN** the `categorization_rules` table and `rule_scope` enum SHALL be dropped

### Requirement: Rule CRUD operations
The rule store SHALL support Create, Update, Delete, and List operations for categorization rules.

#### Scenario: Create rule
- **WHEN** a rule is created
- **THEN** it SHALL be persisted and the created rule SHALL be returned with a generated ID

#### Scenario: Update rule
- **WHEN** a rule is updated
- **THEN** the updated rule SHALL be returned

#### Scenario: Delete rule
- **WHEN** a rule is deleted
- **THEN** it SHALL be removed from the database

#### Scenario: List rules by workspace
- **WHEN** `ListByWorkspace(ctx, wsID)` is called
- **THEN** it SHALL return all rules with workspace scope for that workspace, ordered by priority DESC

### Requirement: Rule enable/disable toggle
The rule store SHALL support toggling a rule's enabled status without modifying other fields.

#### Scenario: Disable an enabled rule
- **WHEN** `ToggleEnabled(ctx, ruleID, false)` is called on an enabled rule
- **THEN** the rule's `enabled` field SHALL be set to false

#### Scenario: Enable a disabled rule
- **WHEN** `ToggleEnabled(ctx, ruleID, true)` is called on a disabled rule
- **THEN** the rule's `enabled` field SHALL be set to true

### Requirement: Rule resolution for auto-categorization
The rule service SHALL resolve the best matching category for a transaction description by fetching enabled rules ordered by scope precedence DESC then priority DESC, and returning the first match using case-insensitive ILIKE pattern matching.

#### Scenario: Workspace rule beats user rule
- **WHEN** both a workspace rule and a user rule match a description
- **THEN** the workspace rule's target category SHALL be returned

#### Scenario: Higher priority wins within same scope
- **WHEN** two workspace rules match, one with priority 10 and one with priority 5
- **THEN** the rule with priority 10's target category SHALL be returned

#### Scenario: No matching rule
- **WHEN** no enabled rule's pattern matches the description
- **THEN** the resolution SHALL return nil (no category suggestion)

#### Scenario: Disabled rules are skipped
- **WHEN** a disabled rule's pattern matches the description but no enabled rule matches
- **THEN** the resolution SHALL return nil

### Requirement: Rule service permission enforcement
Workspace rule CRUD SHALL require editor+ permission. User rule CRUD SHALL require the rule's owner_id to match the caller.

#### Scenario: Create workspace rule requires editor
- **WHEN** a viewer attempts to create a workspace rule
- **THEN** it SHALL return `model.ErrInsufficientPermission`

#### Scenario: Toggle rule requires editor
- **WHEN** an editor toggles a workspace rule's enabled status
- **THEN** the toggle SHALL succeed
