## 1. Funding Model & Migration

- [x] 1.1 Create `model/funding.go` — `FundingID` typed wrapper, `Funding` struct (ID, WorkspaceID, UserID, YearMonth string, Amount Money, CreatedAt, UpdatedAt)
- [x] 1.2 Create `migrations/000005_create_fundings.up.sql` — `fundings` table (BIGSERIAL PK, workspace_id FK ON DELETE CASCADE, user_id FK, year_month TEXT NOT NULL, amount NUMERIC(19,4) NOT NULL, currency TEXT DEFAULT 'PLN', timestamps), unique constraint `(workspace_id, user_id, year_month)`
- [x] 1.3 Create `migrations/000005_create_fundings.down.sql` — drop fundings table

## 2. Rule Model & Migration

- [x] 2.1 Create `model/rule.go` — `RuleID` typed wrapper, `RuleScope` enum (ScopeSystem/ScopeUser/ScopeWorkspace), `CategorizationRule` struct (ID, Scope, OwnerID *UserID, WorkspaceID *WorkspaceID, MatchPattern, TargetCategoryID, Priority int, Enabled bool, timestamps), `ScopePrecedence()` method (workspace=3, user=2, system=1)
- [x] 2.2 Create `migrations/000006_create_categorization_rules.up.sql` — `rule_scope` enum, `categorization_rules` table (BIGSERIAL PK, scope, owner_id nullable FK, workspace_id nullable FK ON DELETE CASCADE, match_pattern TEXT, target_category_id FK, priority INT DEFAULT 0, enabled BOOLEAN DEFAULT TRUE, timestamps), CHECK constraint for scope-FK consistency, partial indexes
- [x] 2.3 Create `migrations/000006_create_categorization_rules.down.sql` — drop categorization_rules, rule_scope enum

## 3. Queries & sqlc

- [x] 3.1 Create `db/queries/fundings.sql` — sqlc queries: UpsertFunding (INSERT ON CONFLICT UPDATE), ListFundingsByWorkspace (with year_month range), GetFundingsByWorkspaceAndUsers (for summary integration)
- [x] 3.2 Create `db/queries/rules.sql` — sqlc queries: InsertRule, UpdateRule, DeleteRule, ListRulesByWorkspace, ToggleRuleEnabled, ListEnabledRulesForResolution (ordered by scope precedence DESC, priority DESC, enabled=true only)
- [x] 3.3 Run `sqlc generate` to update generated code in `store/`

## 4. Stores

- [x] 4.1 Create `store/funding_store.go` — `FundingStore` with pool dependency. Methods: Upsert, ListByWorkspace (year_month range), ListByWorkspaceAndUsers (for summary). Model mapping with pgtype conversions.
- [x] 4.2 Create `store/rule_store.go` — `RuleStore` with pool dependency. Methods: Create, Update, Delete, ListByWorkspace, ToggleEnabled, ResolveForDescription (fetches enabled rules in precedence order, applies ILIKE matching in Go or SQL). Model mapping.

## 5. Services

- [x] 5.1 Create `service/funding_service.go` — `FundingService` with funding store, workspace store. Co-located DTOs: RecordFundingInput, ListFundingsInput. Methods: Record (editor+, upserts), ListByRange (viewer+)
- [x] 5.2 Create `service/rule_service.go` — `RuleService` with rule store, workspace store. Co-located DTOs: CreateRuleInput, UpdateRuleInput. Methods: Create (editor+), Update (editor+), Delete (editor+), Toggle (editor+), ListByWorkspace (viewer+), ResolveCategory(ctx, wsID, userID, description) → *CategoryID
- [x] 5.3 Update `service/summary_service.go` — add FundingStore dependency to SummaryService, query funding by overlapping months, merge funded amounts into ParticipantSpending, compute Balance = Funded - Spent
- [x] 5.4 Update `service/transaction_service.go` — add RuleService dependency to TransactionService, call ResolveCategory during ImportTransactions for auto-categorization

## 6. HTTP Handlers

- [x] 6.1 Create `handler/handler_funding.go` — `FundingHandler` struct wrapping service. Methods: RecordFunding (PUT), ListFundings (GET with from/to month params). Parse year_month from body and query params.
- [x] 6.2 Create `handler/handler_rule.go` — `RuleHandler` struct wrapping service. Methods: CreateRule (POST), UpdateRule (PUT), DeleteRule (DELETE), ToggleRule (PATCH). Parse workspace ID, rule ID from URL params.

## 7. Router Wiring

- [x] 7.1 Update `handler/router.go` — add FundingHandler and RuleHandler to RouterConfig, register routes: `PUT /fundings`, `GET /fundings`, `POST /rules`, `PUT /rules/{ruleID}`, `DELETE /rules/{ruleID}`, `PATCH /rules/{ruleID}/toggle`
- [x] 7.2 Update `cmd/api/main.go` — create FundingStore, RuleStore, FundingService, RuleService, FundingHandler, RuleHandler, update SummaryService constructor with FundingStore, update TransactionService constructor with RuleService, pass handlers to NewRouter

## 8. Error Sentinels

- [x] 8.1 Update `model/errors.go` — add `ErrFundingNotFound`, `ErrRuleNotFound` sentinel errors

## 9. Verification

- [x] 9.1 Run `go build ./...` — must compile with zero errors
- [x] 9.2 Run `go vet ./...` — must pass with zero warnings
- [x] 9.3 Run `go test -count=1 ./...` — all tests must pass
