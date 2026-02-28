## ADDED Requirements

### Requirement: Workspace validation tests
Tests SHALL verify `Workspace.Validate()` returns the correct sentinel error for each invalid state and nil for valid input. Tests SHALL cover: empty name, empty description, invalid type, and a fully valid workspace.

#### Scenario: Empty workspace name
- **WHEN** `Validate()` is called on a Workspace with an empty Name
- **THEN** it returns `model.ErrWorkspaceNameRequired`

#### Scenario: Empty workspace description
- **WHEN** `Validate()` is called on a Workspace with a valid Name but empty Description
- **THEN** it returns `model.ErrWorkspaceDescriptionRequired`

#### Scenario: Invalid workspace type
- **WHEN** `Validate()` is called on a Workspace with Type not equal to `TypePrivate` or `TypeShared`
- **THEN** it returns `model.ErrInvalidWorkspaceType`

#### Scenario: Valid workspace
- **WHEN** `Validate()` is called on a Workspace with valid Name, Description, and Type
- **THEN** it returns nil

### Requirement: MemberRole level tests
Tests SHALL verify `MemberRole.Level()` returns the correct numeric level for all defined roles and the zero value for unknown roles. Table-driven with cases: owner=3, editor=2, viewer=1, unknown=0.

#### Scenario: All defined roles
- **WHEN** `Level()` is called on each of `RoleOwner`, `RoleEditor`, `RoleViewer`
- **THEN** it returns 3, 2, 1 respectively

#### Scenario: Unknown role
- **WHEN** `Level()` is called on `MemberRole("invalid")`
- **THEN** it returns 0

### Requirement: WorkspaceMember permission tests
Tests SHALL verify `CanEdit()` and `CanView()` for all role levels. Owner and Editor can edit; all roles can view; unknown role cannot view or edit.

#### Scenario: Owner can edit and view
- **WHEN** `CanEdit()` and `CanView()` are called on a member with `RoleOwner`
- **THEN** both return true

#### Scenario: Editor can edit and view
- **WHEN** `CanEdit()` and `CanView()` are called on a member with `RoleEditor`
- **THEN** both return true

#### Scenario: Viewer cannot edit but can view
- **WHEN** `CanEdit()` and `CanView()` are called on a member with `RoleViewer`
- **THEN** `CanEdit()` returns false, `CanView()` returns true

#### Scenario: Unknown role cannot edit or view
- **WHEN** `CanEdit()` and `CanView()` are called on a member with `MemberRole("unknown")`
- **THEN** both return false

### Requirement: Transaction ValidateEntries tests
Tests SHALL verify `Transaction.ValidateEntries()` for: no entries, entry sum mismatch, and valid entries summing to total.

#### Scenario: No entries
- **WHEN** `ValidateEntries()` is called on a Transaction with an empty Entries slice
- **THEN** it returns `model.ErrTransactionNoEntries`

#### Scenario: Entry sum does not match total
- **WHEN** `ValidateEntries()` is called on a Transaction where entry amounts sum to a value different from TotalAmount
- **THEN** it returns `model.ErrEntrySumMismatch`

#### Scenario: Valid entries
- **WHEN** `ValidateEntries()` is called on a Transaction where entry amounts exactly sum to TotalAmount
- **THEN** it returns nil

#### Scenario: Multiple entries summing correctly
- **WHEN** `ValidateEntries()` is called with 3 entries whose amounts sum to TotalAmount
- **THEN** it returns nil

### Requirement: CategorizationRule ScopePrecedence tests
Tests SHALL verify `ScopePrecedence()` returns: workspace=3, user=2, system=1, unknown=0.

#### Scenario: All defined scopes
- **WHEN** `ScopePrecedence()` is called on rules with `ScopeWorkspace`, `ScopeUser`, `ScopeSystem`
- **THEN** it returns 3, 2, 1 respectively

#### Scenario: Unknown scope
- **WHEN** `ScopePrecedence()` is called on a rule with `RuleScope("other")`
- **THEN** it returns 0

### Requirement: Sentinel error identity tests
Tests SHALL verify that all sentinel errors in `model/errors.go` are distinct and can be matched with `errors.Is()`. Tests SHALL confirm wrapping preserves identity via `fmt.Errorf("context: %w", err)`.

#### Scenario: Direct identity check
- **WHEN** `errors.Is(model.ErrWorkspaceNotFound, model.ErrWorkspaceNotFound)` is evaluated
- **THEN** it returns true

#### Scenario: Wrapped error preserves identity
- **WHEN** an error is wrapped with `fmt.Errorf("ctx: %w", model.ErrWorkspaceNotFound)` and checked with `errors.Is()`
- **THEN** it matches the original sentinel error

#### Scenario: Different sentinels are distinct
- **WHEN** `errors.Is(model.ErrWorkspaceNotFound, model.ErrUserNotFound)` is evaluated
- **THEN** it returns false
