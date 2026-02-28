## Why

The SpendWise backend has only 4 test files covering ~8% of 51+ source files. All business logic (services), HTTP handlers, and data access layers are completely untested. Bugs in permission checks, transaction validation, budget limit enforcement, and rule-based categorization can go undetected. Adding a comprehensive test suite now — before the codebase grows further — prevents regressions, documents expected behavior, and enables safe refactoring.

## What Changes

- Add **model-layer unit tests** for all untested domain types: `Workspace.Validate()`, `MemberRole.Level()`, `WorkspaceMember.CanEdit()/CanView()`, `Transaction.ValidateEntries()`, `CategorizationRule.ScopePrecedence()`, and sentinel error identity checks
- Introduce **service-layer interfaces** for store dependencies to enable mock-based testing without a database, then add unit tests for all 6 service files: `WorkspaceService`, `TransactionService`, `RuleService`, `FundingService`, `SummaryService`, `AuthService`
- Add **handler-layer tests** using `httptest` for all 7 handler files, covering request decoding, response format, status codes, and error mapping
- Add tests for **auth gaps**: `context.go` (round-trip context storage), `dev_bypass.go` (middleware behavior)
- Add **config tests**: `ServerConfig.Addr()`, `DatabaseConfig.DSN()`, `Load()` defaults
- Add **shared test helpers** (`testutil/` or helpers within packages) for common operations: creating test Money values, building authenticated requests, asserting JSON responses
- Establish **testing standards document** within specs: table-driven tests, `t.Helper()`, `got/want` error messages, `testing.TB` in helpers, `-race` compatibility

## Capabilities

### New Capabilities
- `testing-standards`: Go testing conventions, patterns, and quality bar for all SpendWise tests (table-driven, helpers, mocking strategy, error messages, race-safety)
- `model-tests`: Unit tests for all model-layer types — validation methods, role level comparisons, entry sum validation, scope precedence, sentinel error identity
- `service-tests`: Unit tests for all service-layer business logic — permission enforcement, budget validation, transaction import/dedup, rule resolution, category operations — using interface-based mocks for store dependencies
- `handler-tests`: HTTP handler tests using `httptest` — request decoding, JSON response format, status code mapping, auth context extraction, error envelope consistency
- `auth-config-tests`: Tests for auth context helpers, dev bypass middleware, and config struct methods/defaults

### Modified Capabilities

## Impact

- **Code changes**: New `_test.go` files across `model/`, `service/`, `handler/`, `auth/`, `config/` packages. Possible introduction of store interfaces or a `testutil` package for shared helpers.
- **Dependencies**: No new external dependencies (use stdlib `testing`, `httptest`, `io`). May optionally add `github.com/google/go-cmp` for struct diff assertions.
- **CI**: Tests run via existing `make test` (`go test -race -count=1 ./...`). No pipeline changes needed.
- **Risk**: Service tests require introducing interfaces for `store.*Store` types (currently concrete struct dependencies). This is a refactor that improves testability but touches constructor signatures across `service/` and `cmd/api/main.go`.
