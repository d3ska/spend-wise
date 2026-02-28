## Why

In shared workspaces, participants contribute different amounts to the household pool. Without funding tracking, there's no way to answer "who owes whom?" — the core collaboration question. Categorization rules automate the tedious task of assigning categories to imported transactions, following a tiered hierarchy (workspace rules > user rules > system defaults) so households can customize without losing global defaults.

**Depends on:** `spending-summary` (summary service integrates funding data into participant balances)

## What Changes

### Funding (participant contributions)

- Create `Funding` domain entity with typed `FundingID`, workspace-scoped, user-scoped, monthly granularity (year_month), `Money` amount
- Create `domain.FundingRepository` interface (Upsert, ListByWorkspaceAndMonth)
- Create database migration `000005_create_fundings` with unique constraint `(workspace_id, user_id, year_month)`
- Create sqlc queries (`db/queries/fundings.sql`)
- Create `adapter/postgres/funding_repo.go` with compile-time interface check
- Create `app/funding_service.go` with co-located DTOs:
  - `Record` — verifies editor+ permission, upserts funding for user/month (idempotent)
  - `ListByRange` — returns fundings for a workspace filtered by date range
- Create `adapter/httpapi/handler_funding.go`:
  - `PUT /api/v1/workspaces/{id}/fundings` — record/update a funding contribution (upsert)
  - `GET /api/v1/workspaces/{id}/fundings?from=&to=` — list fundings in range
- Update `SummaryService` to include funding data in `ParticipantSpending.Funded` and `ParticipantSpending.Balance`

### Categorization Rules (tiered automation)

- Create `CategorizationRule` domain entity with typed `RuleID`, `RuleScope` (system|user|workspace), nullable owner_id (user scope), nullable workspace_id (workspace scope), match_pattern, target_category_id, priority, `ScopePrecedence()` method
- Create `domain.RuleRepository` interface (Create, Update, Delete, ResolveForTransaction — returns rules ordered by scope precedence DESC, priority DESC)
- Create database migration `000006_create_categorization_rules` with:
  - `rule_scope` PostgreSQL enum
  - CHECK constraint enforcing scope↔FK consistency (user scope requires owner_id, workspace scope requires workspace_id)
  - Partial indexes on workspace_id and owner_id
- Create sqlc queries (`db/queries/rules.sql`)
- Create `adapter/postgres/rule_repo.go` with compile-time interface check
- Create `app/rule_service.go` with co-located DTOs:
  - `Create`, `Update`, `Delete` — CRUD with permission checks (workspace rules require editor+, user rules require matching user)
  - `ResolveCategory(ctx, wsID, userID, description) → *CategoryID` — fetches rules in precedence order, applies pattern matching, returns first match or nil
- Create `adapter/httpapi/handler_rule.go`:
  - `POST /api/v1/workspaces/{id}/rules` — create rule
  - `PUT /api/v1/workspaces/{id}/rules/{ruleID}` — update rule
  - `DELETE /api/v1/workspaces/{id}/rules/{ruleID}` — delete rule
- Update `TransactionService.Import` to call `RuleService.ResolveCategory` for auto-categorization during bank import
- Wire funding and rule services/handlers into router

## Capabilities

### New Capabilities
- `funding-tracking`: Participant contribution recording with monthly granularity, upsert semantics, date-range listing. Integrates with spending summary to compute per-participant balances (funded - spent).
- `categorization-rules`: Tiered auto-categorization rules (workspace > user > system precedence), pattern matching against transaction descriptions, CRUD with scope-appropriate permission checks, automatic application during bank imports.

### Modified Capabilities
- `spending-summary`: ParticipantSpending now includes Funded amount and Balance (Funded - Spent)
- `bank-import`: Import now auto-categorizes transactions using the tiered rule resolution engine
- `http-server`: Add funding and rule routes to the Chi router

## Impact

- **API:** 5 new endpoints — funding (2) + rules (3)
- **Database:** Migrations 005-006 create `fundings` and `categorization_rules` tables
- **Cross-cutting:** This change modifies two existing services (SummaryService, TransactionService) to integrate funding data and auto-categorization — the first change with cross-service dependencies
- **Code:** ~8 new files (2 domain, 2 repo, 2 service, 2 handler) + modifications to 2 existing services
