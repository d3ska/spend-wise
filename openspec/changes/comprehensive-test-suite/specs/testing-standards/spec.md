## ADDED Requirements

### Requirement: Table-driven test structure
All tests with multiple cases SHALL use table-driven tests with a slice or map of anonymous structs. Each test case SHALL have a descriptive `name` field. Each case SHALL be executed via `t.Run(name, ...)` to enable selective execution and clear failure output.

#### Scenario: Multiple input variations
- **WHEN** a function is tested with more than one input/output combination
- **THEN** all cases are expressed as entries in a single `[]struct{...}` or `map[string]struct{...}` and iterated with `t.Run()`

#### Scenario: Selective test execution
- **WHEN** a developer runs `go test -run TestFoo/case_name`
- **THEN** only the matching subtest executes, because all cases use `t.Run()` with descriptive names

### Requirement: Error message format
All test failure messages SHALL include both the actual ("got") and expected ("want") values. Messages SHALL use `t.Errorf("got %v, want %v", got, want)` or `t.Errorf("got %q, want %q", got, want)` for strings.

#### Scenario: Assertion failure output
- **WHEN** a test assertion fails
- **THEN** the error message contains both the actual and expected values in `got X, want Y` format

### Requirement: Test helper functions use t.Helper
All test helper functions that call `t.Errorf`, `t.Fatalf`, or `t.Fatal` SHALL call `t.Helper()` as their first statement. Helper functions that accept a testing parameter SHALL accept `testing.TB` (not `*testing.T`) to support both tests and benchmarks.

#### Scenario: Helper failure reporting
- **WHEN** an assertion helper reports a failure
- **THEN** the failure points to the caller's line (not the helper's line) because `t.Helper()` was called

### Requirement: Race-safe tests
All tests SHALL pass when run with `go test -race`. Tests MUST NOT use shared mutable state across goroutines without synchronization. Parallel subtests SHALL use `t.Parallel()` only when test cases are truly independent.

#### Scenario: Race detector clean run
- **WHEN** `go test -race ./...` is executed
- **THEN** zero data race warnings are reported

### Requirement: Errorf vs Fatalf usage
Tests SHALL use `t.Errorf` for assertions where subsequent checks are still meaningful. Tests SHALL use `t.Fatalf` only when a failure makes subsequent checks invalid (e.g., nil pointer would panic).

#### Scenario: Non-blocking assertion
- **WHEN** an equality check fails but subsequent checks are independent
- **THEN** `t.Errorf` is used so all failures are reported in one run

#### Scenario: Precondition failure
- **WHEN** an error return is non-nil and the result value would be meaningless to check
- **THEN** `t.Fatalf` is used to stop the subtest immediately

### Requirement: No external test frameworks
Tests SHALL use only the Go standard library (`testing`, `net/http/httptest`, `io`, `bytes`, `encoding/json`) and optionally `github.com/google/go-cmp` for struct diff comparisons. Tests SHALL NOT introduce testify, gomock, or other third-party test frameworks.

#### Scenario: Dependency check
- **WHEN** reviewing test file imports
- **THEN** no imports of `github.com/stretchr/testify` or `github.com/golang/mock` are present

### Requirement: Mock pattern using function fields
Service-layer mocks SHALL use structs with function fields (e.g., `CreateFunc func(...) (...)`) that delegate to the configured function. This allows per-test-case behavior configuration without a mock framework.

#### Scenario: Configurable mock behavior
- **WHEN** a service test needs a store mock that returns a specific error
- **THEN** the test sets the mock's function field inline (e.g., `mock.GetMemberFunc = func(...) { return ..., model.ErrNotWorkspaceMember }`)
