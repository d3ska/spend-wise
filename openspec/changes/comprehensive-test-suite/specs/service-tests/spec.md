## ADDED Requirements

### Requirement: Store interfaces for testability
Each service SHALL depend on a narrow interface (defined in the service package) instead of a concrete `*store.XxxStore` pointer. The interface SHALL include only the methods that service actually calls. Concrete store types SHALL satisfy these interfaces without modification.

#### Scenario: WorkspaceService store interface
- **WHEN** `WorkspaceService` is constructed
- **THEN** it accepts a `WorkspaceStoreIface` and `CategoryStoreIface` interface, not concrete store types

#### Scenario: Concrete store satisfies interface
- **WHEN** `*store.WorkspaceStore` is passed where `WorkspaceStoreIface` is expected
- **THEN** compilation succeeds because the concrete type already implements all required methods

### Requirement: WorkspaceService permission enforcement tests
Tests SHALL verify that every WorkspaceService method enforces the correct minimum role. Tests SHALL cover: allowed role succeeds, insufficient role returns `ErrInsufficientPermission`, non-member returns `ErrNotWorkspaceMember`.

#### Scenario: CreateWorkspace with valid input
- **WHEN** `CreateWorkspace` is called by a user with valid input
- **THEN** the workspace is created and the user is added as owner

#### Scenario: CreateWorkspace with invalid name
- **WHEN** `CreateWorkspace` is called with an empty name
- **THEN** it returns `ErrWorkspaceNameRequired`

#### Scenario: GetWorkspace as viewer
- **WHEN** `GetWorkspace` is called by a user with viewer role
- **THEN** the workspace is returned successfully

#### Scenario: GetWorkspace as non-member
- **WHEN** `GetWorkspace` is called by a user who is not a member
- **THEN** it returns `ErrNotWorkspaceMember`

#### Scenario: UpdateWorkspace requires editor role
- **WHEN** `UpdateWorkspace` is called by a viewer
- **THEN** it returns `ErrInsufficientPermission`

#### Scenario: DeleteWorkspace requires owner role
- **WHEN** `DeleteWorkspace` is called by an editor
- **THEN** it returns `ErrInsufficientPermission`

#### Scenario: AddMember requires owner role
- **WHEN** `AddMember` is called by an editor
- **THEN** it returns `ErrInsufficientPermission`

#### Scenario: DeleteCategory blocks Uncategorized
- **WHEN** `DeleteCategory` is called for a category named "Uncategorized"
- **THEN** it returns `ErrCategoryUndeletable`

### Requirement: TransactionService tests
Tests SHALL verify transaction creation, validation, import/dedup logic, and rule application using mock stores.

#### Scenario: Create transaction with valid entries
- **WHEN** `Create` is called with entries summing to the total amount
- **THEN** the transaction is persisted and returned

#### Scenario: Create transaction with mismatched entries
- **WHEN** `Create` is called with entries that do not sum to the total amount
- **THEN** it returns `ErrEntrySumMismatch`

#### Scenario: Create defaults type to expense
- **WHEN** `Create` is called with empty Type
- **THEN** the transaction is created with `TypeExpense`

#### Scenario: Create resolves zero CategoryID to Uncategorized
- **WHEN** `Create` is called with an entry whose CategoryID is 0
- **THEN** the entry's CategoryID is set to the Uncategorized category's ID

#### Scenario: Import skips duplicate fingerprint
- **WHEN** `ImportTransactions` is called with a fingerprint that already exists
- **THEN** the transaction is skipped and `ImportResult.Skipped` is incremented

#### Scenario: Import creates new transaction
- **WHEN** `ImportTransactions` is called with a new fingerprint and valid entries
- **THEN** the transaction is created and `ImportResult.Imported` is incremented

#### Scenario: Import auto-categorizes via rule service
- **WHEN** `ImportTransactions` is called and the rule service resolves a category
- **THEN** entries with CategoryID 0 get the resolved category ID

#### Scenario: ApplyRules re-categorizes entries
- **WHEN** `ApplyRules` is called and rules match some entries
- **THEN** matched entries get the rule's category, unmatched entries get Uncategorized

#### Scenario: ApplyRules skips transfer entries
- **WHEN** `ApplyRules` encounters an entry with `TypeTransfer`
- **THEN** that entry is skipped and not re-categorized

#### Scenario: UpdateTransaction checks workspace ownership
- **WHEN** `UpdateTransaction` is called with a transaction from a different workspace
- **THEN** it returns `ErrTransactionNotFound`

### Requirement: RuleService permission tests
Tests SHALL verify that rule CRUD operations enforce editor+ permission and that listing requires viewer+.

#### Scenario: Create rule as editor
- **WHEN** `Create` is called by an editor
- **THEN** the rule is created with workspace scope and Enabled=true

#### Scenario: Create rule as viewer
- **WHEN** `Create` is called by a viewer
- **THEN** it returns `ErrInsufficientPermission`

#### Scenario: ListByWorkspace as viewer
- **WHEN** `ListByWorkspace` is called by a viewer
- **THEN** the rules are returned successfully

#### Scenario: ResolveCategory delegates to store
- **WHEN** `ResolveCategory` is called
- **THEN** it delegates to `ruleStore.ResolveForDescription` and returns the result

### Requirement: FundingService budget validation tests
Tests SHALL verify budget limit enforcement: category budgets cannot exceed overall budget.

#### Scenario: Category budget within overall limit
- **WHEN** a category budget is recorded and the sum of all category budgets is within the overall budget
- **THEN** the funding is persisted successfully

#### Scenario: Category budget exceeds overall limit
- **WHEN** a category budget would cause the sum of category budgets to exceed the overall budget
- **THEN** it returns `ErrCategoryBudgetsExceed`

#### Scenario: Overall budget below existing category sum
- **WHEN** the overall budget is set to a value below the sum of existing category budgets
- **THEN** it returns `ErrCategoryBudgetsExceed`

#### Scenario: Category budget with no overall budget set
- **WHEN** a category budget is recorded and no overall budget exists
- **THEN** the funding is persisted successfully (no constraint to enforce)

#### Scenario: Record funding requires editor role
- **WHEN** `Record` is called by a viewer
- **THEN** it returns `ErrInsufficientPermission`

### Requirement: SummaryService tests
Tests SHALL verify permission enforcement and correct aggregation of funding data into participant spending.

#### Scenario: GetSummary as viewer
- **WHEN** `GetSummary` is called by a viewer
- **THEN** the summary is returned with total spent, by-category, and by-participant data

#### Scenario: GetSummary merges funding into participant spending
- **WHEN** `GetSummary` is called and funding data exists for participants
- **THEN** each participant's Funded and Balance fields are populated

#### Scenario: GetSummary as non-member
- **WHEN** `GetSummary` is called by a non-member
- **THEN** it returns `ErrNotWorkspaceMember`

### Requirement: AuthService tests
Tests SHALL verify the OAuth flow orchestration using mock provider and mock user store.

#### Scenario: Successful authentication
- **WHEN** `Authenticate` is called with a valid provider and code
- **THEN** it returns a signed JWT and the user

#### Scenario: Unsupported provider
- **WHEN** `Authenticate` is called with an unknown provider name
- **THEN** it returns an error containing "unsupported provider"

#### Scenario: Exchange failure
- **WHEN** the OAuth provider's `Exchange` returns an error
- **THEN** `Authenticate` returns a wrapped error

#### Scenario: FetchUser failure
- **WHEN** the OAuth provider's `FetchUser` returns an error
- **THEN** `Authenticate` returns a wrapped error
