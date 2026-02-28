## Context

SpendWise backend has 51+ Go source files but only 4 test files (~8% coverage). The tested parts are low-level primitives — `Money`, `DateRange` value objects and `JWT`/`middleware` auth. All business logic (6 services), HTTP handlers (7 files), and data access (8 stores) are untested.

Services currently take concrete `*store.XxxStore` pointers, making it impossible to unit-test business logic without a running PostgreSQL instance. The existing test style uses table-driven tests with `t.Run()`, `got/want` error messages, and small helpers — this design preserves and extends those conventions.

**Constraints:**
- Must not break existing tests or API behavior
- Must not require a database for unit tests (service & handler layers)
- Must run cleanly with `go test -race -count=1 ./...`
- Store integration tests are out of scope (they need a test DB and are a separate change)

## Goals / Non-Goals

**Goals:**
- Define store interfaces so service-layer tests can use in-memory mocks
- Establish shared test helpers and patterns for reuse across packages
- Achieve unit test coverage for all model validation, service business logic, handler HTTP behavior, and config methods
- All tests follow existing project conventions: table-driven, `t.Helper()`, `got/want` messages

**Non-Goals:**
- Database integration tests for the store layer (separate change — requires test DB setup, migrations, teardown)
- End-to-end / acceptance tests (requires running server + DB)
- Bank sync integration tests (requires external API mocking at HTTP transport level)
- Achieving a specific coverage percentage target (focus on meaningful tests, not metrics)
- Adding third-party test frameworks (testify, gomock) — use stdlib only, optionally `go-cmp`

## Decisions

### 1. Introduce store interfaces in the service package (not the store package)

**Decision:** Define narrow interfaces in each service file (or a shared `service/iface.go`) that match only the store methods that service actually calls. Services accept these interfaces instead of concrete `*store.XxxStore` pointers.

**Rationale:** Go convention is "accept interfaces, return structs" and interfaces belong to the consumer. Narrow interfaces (3-6 methods each) are easy to mock manually. The concrete `*store.XxxStore` already satisfies these interfaces without modification.

**Alternatives considered:**
- *Generate mocks from full store structs via gomock* — adds a code-gen dependency, interfaces would be ~10 methods each (too wide)
- *Use the sqlc-generated `Querier` interface (61 methods)* — too broad, couples tests to DB-level details
- *Test services with a real DB* — too slow for unit tests, belongs in integration test scope

### 2. Manual mock structs, no mock framework

**Decision:** Write mock structs by hand in `_test.go` files. Each mock uses function fields (`XxxFunc func(...)`) so test cases can configure behavior inline.

**Rationale:** The project has no dependency on testify/gomock. Manual mocks are transparent, require no code generation, and match the project's minimalist dependency philosophy. The interface surfaces are small (3-6 methods per service) making manual mocks practical.

**Example pattern:**
```go
type mockWorkspaceStore struct {
    GetMemberFunc func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error)
    CreateFunc    func(ctx context.Context, ws model.Workspace) (model.Workspace, error)
    // ...
}
func (m *mockWorkspaceStore) GetMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
    return m.GetMemberFunc(ctx, wsID, userID)
}
```

### 3. Handler tests use httptest directly (no router)

**Decision:** Test handlers by calling `handler.Method(recorder, request)` directly, using `httptest.NewRequest` and `httptest.NewRecorder`. For chi URL params, set them via `chi.NewRouteContext()`.

**Rationale:** Testing handlers in isolation (without the full router) keeps tests focused on one handler's behavior. The router itself is thin (just wiring) and doesn't need separate tests. Chi URL params can be injected into the request context.

**Alternatives considered:**
- *Spin up the full router per test* — slower, tests become integration tests, failures harder to isolate
- *Use a test server (`httptest.NewServer`)* — overkill for unit tests, useful for acceptance tests later

### 4. Auth context injection via helper function

**Decision:** Create a test helper `withTestUser(ctx, userID)` that wraps `auth.withUserID` (which is unexported). Since handler tests are in the `handler` package but need to set auth context, the helper will use `auth.UserIDFromContext` for verification and the middleware's public API for setup.

**Rationale:** Handlers extract user ID via `auth.UserIDFromContext(r.Context())`. Tests need to set this. Since `auth.withUserID` is unexported, handler tests will either: (a) export a test-only helper from auth, or (b) use the existing middleware with a valid JWT. Option (a) is cleaner — add an exported `auth.TestContext(ctx, userID)` guarded by a `_test.go` file in the auth package.

### 5. No new packages — helpers live in test files

**Decision:** Test helpers go in `_test.go` files within their package. Cross-package helpers (if any) can go in an `internal/testutil` package, but only if genuinely shared across 3+ packages.

**Rationale:** Avoid premature abstraction. Most helpers are package-specific (e.g., `dec()` in model tests, mock structs in service tests). Creating a `testutil` package for 1-2 helpers adds unnecessary structure.

### 6. Optional `go-cmp` for struct comparisons

**Decision:** Add `github.com/google/go-cmp` as a test-only dependency for comparing complex structs in service and handler tests. Use `cmp.Diff()` for readable failure output.

**Rationale:** The model layer has many structs with multiple fields. `reflect.DeepEqual` gives poor error messages. `cmp.Diff()` shows exactly which field differs. This is the Go community's recommended approach (Dave Cheney, Go wiki). It's a single, well-maintained dependency with no transitive deps.

## Risks / Trade-offs

**[Risk] Introducing interfaces changes service constructor signatures** → Mitigation: The concrete store types already satisfy the new interfaces. Only constructor parameter types change (from `*store.XxxStore` to the interface). `cmd/api/main.go` passes concrete stores which auto-satisfy. No runtime behavior change.

**[Risk] Manual mocks become tedious if interfaces grow** → Mitigation: Interfaces are scoped to what each service actually uses (3-6 methods). If they grow beyond ~8 methods, revisit with code generation. This is unlikely given the architecture.

**[Risk] Handler tests may be brittle to response format changes** → Mitigation: Test status codes and key JSON fields, not exact byte-for-byte response bodies. Use struct decoding for assertions.

**[Risk] `auth.TestContext` export leaks test infrastructure** → Mitigation: Place it in a `_test.go` file within the `auth` package (e.g., `auth/export_test.go`), so it's only available to test code. Alternatively, export it in a file with a `//go:build testing` build tag.

## Open Questions

- Should `go-cmp` be added now or deferred until a test actually needs deep struct comparison? (Recommendation: add it — service tests will need it immediately for comparing `model.Workspace`, `model.Transaction` etc.)
