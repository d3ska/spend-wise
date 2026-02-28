## 1. Infrastructure & Dependencies

- [x] 1.1 Add `github.com/google/go-cmp` as a test dependency (`go get -t github.com/google/go-cmp/cmp`)
- [x] 1.2 Create `auth/export_test.go` with exported `WithTestUserID(ctx, userID)` helper so handler tests can inject auth context
- [x] 1.3 Create `handler/testhelpers_test.go` with helpers: `newAuthenticatedRequest(method, path, body, userID)`, `withChiParam(r, key, value)`, `decodeResponseBody(t, w, dst)`

## 2. Model Layer Tests

- [x] 2.1 Create `model/workspace_test.go` — table-driven tests for `Workspace.Validate()` (empty name, empty description, invalid type, valid)
- [x] 2.2 Add to `model/workspace_test.go` — tests for `MemberRole.Level()` (owner=3, editor=2, viewer=1, unknown=0)
- [x] 2.3 Add to `model/workspace_test.go` — tests for `WorkspaceMember.CanEdit()` and `CanView()` across all roles
- [x] 2.4 Create `model/transaction_test.go` — tests for `Transaction.ValidateEntries()` (no entries, sum mismatch, valid single entry, valid multiple entries)
- [x] 2.5 Create `model/rule_test.go` — tests for `CategorizationRule.ScopePrecedence()` (workspace=3, user=2, system=1, unknown=0)
- [x] 2.6 Create `model/errors_test.go` — tests for sentinel error identity: direct match, wrapped match via `errors.Is()`, distinct errors are not equal
- [x] 2.7 Run `go test -race ./model/...` and verify all pass

## 3. Service Layer — Interfaces & Mocks

- [x] 3.1 Define `WorkspaceStoreIface` interface in `service/workspace_service.go` with methods: `Create`, `GetByID`, `ListByUser`, `Update`, `Delete`, `AddMember`, `GetMember`, `ListMembers`, `RemoveMember` — update `WorkspaceService` struct and `NewWorkspaceService` to accept this interface
- [x] 3.2 Define `CategoryStoreIface` interface in `service/workspace_service.go` with methods: `Create`, `GetByID`, `ListByWorkspace`, `Update`, `GetByName`, `Delete` — update `WorkspaceService` and `TransactionService` to accept this interface
- [x] 3.3 Define `TransactionStoreIface` interface in `service/transaction_service.go` with methods: `Create`, `GetByID`, `ListByWorkspace`, `Update`, `Delete`, `ExistsByFingerprint`, `ListAllEntries`, `UpdateEntryCategory` — update `TransactionService` to accept this interface
- [x] 3.4 Define `RuleStoreIface` interface in `service/rule_service.go` with methods: `Create`, `Update`, `Delete`, `ListByWorkspace`, `ToggleEnabled`, `ResolveForDescription` — update `RuleService` to accept this interface
- [x] 3.5 Define `FundingStoreIface` interface in `service/funding_service.go` with methods: `Upsert`, `ListByWorkspace`, `Delete` — update `FundingService` to accept this interface
- [x] 3.6 Define `SummaryStoreIface` and `FundingStoreIfaceForSummary` interfaces in `service/summary_service.go` — update `SummaryService` to accept these interfaces
- [x] 3.7 Define `UserStoreIface` interface in `service/auth_service.go` with method: `GetOrCreateByEmail` — update `AuthService` to accept this interface
- [x] 3.8 Update `cmd/api/main.go` to pass concrete store types (which already satisfy the interfaces) — verify compilation
- [x] 3.9 Run `go vet ./...` and `go build ./...` to verify no breakage

## 4. Service Layer — WorkspaceService Tests

- [x] 4.1 Create `service/workspace_service_test.go` with mock structs: `mockWorkspaceStore`, `mockCategoryStore` using function-field pattern
- [x] 4.2 Add tests for `CreateWorkspace` — valid input, validation failure (empty name), store error propagation
- [x] 4.3 Add tests for `GetWorkspace` — viewer allowed, non-member returns `ErrNotWorkspaceMember`
- [x] 4.4 Add tests for `UpdateWorkspace` — editor allowed, viewer returns `ErrInsufficientPermission`, validation failure
- [x] 4.5 Add tests for `DeleteWorkspace` — owner allowed, editor returns `ErrInsufficientPermission`
- [x] 4.6 Add tests for `AddMember` and `RemoveMember` — owner allowed, non-owner returns `ErrInsufficientPermission`
- [x] 4.7 Add tests for `CreateCategory`, `ListCategories`, `UpdateCategory`, `DeleteCategory` — permission checks + Uncategorized protection
- [x] 4.8 Run `go test -race ./service/...` on workspace tests

## 5. Service Layer — TransactionService Tests

- [x] 5.1 Create `service/transaction_service_test.go` with mock structs: `mockTransactionStore`, reuse category/workspace mocks
- [x] 5.2 Add tests for `Create` — valid entries, mismatched entries, default type to expense, resolves zero CategoryID to Uncategorized
- [x] 5.3 Add tests for `GetTransaction` and `ListTransactions` — permission checks (viewer allowed, non-member rejected)
- [x] 5.4 Add tests for `DeleteTransaction` — editor allowed, viewer returns `ErrInsufficientPermission`
- [x] 5.5 Add tests for `UpdateTransaction` — editor allowed, workspace mismatch returns `ErrTransactionNotFound`, entry validation
- [x] 5.6 Add tests for `ImportTransactions` — skips duplicate fingerprint, creates new, auto-categorizes via rule service
- [x] 5.7 Add tests for `ApplyRules` — re-categorizes matched entries, sets unmatched to Uncategorized, skips transfers
- [x] 5.8 Run `go test -race ./service/...` on transaction tests

## 6. Service Layer — Rule, Funding, Summary, Auth Tests

- [x] 6.1 Create `service/rule_service_test.go` — mock structs, tests for Create/Update/Delete/Toggle (editor required), ListByWorkspace (viewer required), ResolveCategory delegation
- [x] 6.2 Create `service/funding_service_test.go` — mock structs, tests for Record (permission + budget validation: within limit, exceeds limit, no overall budget, overall below category sum), ListByRange, Delete
- [x] 6.3 Create `service/summary_service_test.go` — mock structs, tests for GetSummary (permission check, merges funding into participant spending, attaches category budgets)
- [x] 6.4 Create `service/auth_service_test.go` — mock provider and mock user store, tests for Authenticate (success, unsupported provider, exchange failure, fetch user failure)
- [x] 6.5 Run `go test -race ./service/...` and verify all pass

## 7. Handler Layer Tests

- [x] 7.1 Create `handler/request_test.go` — tests for `decodeJSON` (valid JSON, invalid JSON, unknown fields rejected, body exceeding 1MB)
- [x] 7.2 Create `handler/response_test.go` — tests for `respondJSON` (Content-Type header, status code, JSON body) and `respondError` (envelope format)
- [x] 7.3 Create `handler/handler_workspace_test.go` — tests for `handleServiceError` mapping all sentinel errors to correct HTTP status codes (table-driven)
- [x] 7.4 Add to `handler/handler_workspace_test.go` — tests for `CreateWorkspace` (success 201, unauthorized 401, invalid body 400), `GetWorkspace` (success 200, not found 404), `DeleteWorkspace` (success 204), `ListWorkspaces` (returns array)
- [x] 7.5 Create `handler/handler_transaction_test.go` — tests for transaction CRUD endpoints (success, unauthorized, validation errors)
- [x] 7.6 Create `handler/handler_auth_test.go` — tests for logout (clears cookie), GetMe (returns user data or 401)
- [x] 7.7 Create handler tests for remaining endpoints: category, funding, rule, summary, bank — at minimum test success and unauthorized cases
- [x] 7.8 Run `go test -race ./handler/...` and verify all pass

## 8. Auth & Config Tests

- [x] 8.1 Create `auth/context_test.go` — tests for `withUserID`/`UserIDFromContext` round-trip, missing value returns `(0, false)`, nested contexts return innermost ID
- [x] 8.2 Create `auth/dev_bypass_test.go` — tests for DevBypassMiddleware: injects user ID, returns 500 on user store failure, caches user via `sync.Once`
- [x] 8.3 Create `config/config_test.go` — tests for `ServerConfig.Addr()`, `DatabaseConfig.DSN()`, and `Load()` default values
- [x] 8.4 Run `go test -race ./auth/... ./config/...` and verify all pass

## 9. Final Validation

- [x] 9.1 Run full test suite: `go test -race -count=1 ./...` — all tests pass
- [x] 9.2 Run `go vet ./...` — no warnings
- [x] 9.3 Verify no test imports testify or gomock
- [x] 9.4 Spot-check that all test functions use table-driven pattern with `t.Run()` where applicable, and error messages use `got/want` format
