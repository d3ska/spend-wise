## Why

Users need to see their spending at a glance — total spent, breakdown by category, breakdown by participant — for any time period (month, week, day, or custom date range). This is the primary UI view: the dashboard. Without it, users must mentally aggregate individual transactions, which defeats the purpose of the application.

**Depends on:** `transactions` (Transaction and Entry entities, transaction repo with date-range queries)

## What Changes

- Create `domain/summary.go` with read models (computed, never stored):
  - `SpendingSummary` — workspace ID, date range, total spent, per-category breakdown, per-participant breakdown
  - `CategorySpending` — category + spent amount
  - `ParticipantSpending` — user + spent amount + funded amount + balance (funded - spent)
- Create `domain.SummaryRepository` interface with aggregate query methods:
  - `TotalSpent(ctx, wsID, dateRange) → Money`
  - `SpentByCategory(ctx, wsID, dateRange) → []CategorySpending`
  - `SpentByParticipant(ctx, wsID, dateRange) → []ParticipantSpending`
- Create sqlc aggregate queries (`db/queries/summaries.sql`):
  - Spending by category: `SELECT c.id, c.name, c.icon, COALESCE(SUM(e.amount), 0) FROM categories LEFT JOIN entries LEFT JOIN transactions WHERE workspace_id AND date range GROUP BY c.id`
  - Spending by participant: `SELECT e.participant_id, COALESCE(SUM(e.amount), 0) FROM entries JOIN transactions WHERE workspace_id AND date range GROUP BY e.participant_id`
  - Total spent: `SELECT COALESCE(SUM(e.amount), 0) FROM entries JOIN transactions WHERE workspace_id AND date range`
- Create `adapter/postgres/summary_repo.go` implementing `domain.SummaryRepository`
- Create `app/summary_service.go`:
  - `GetSummary(ctx, userID, wsID, dateRange) → SpendingSummary` — verifies viewer+ permission, runs all three aggregate queries, also fetches fundings for the overlapping months, assembles the `SpendingSummary` struct with computed balances
  - No DTOs needed — returns domain read models directly
- Create `adapter/httpapi/handler_summary.go`:
  - `GET /api/v1/workspaces/{id}/summary?from=YYYY-MM-DD&to=YYYY-MM-DD` — parses date range from query params, calls service, returns JSON
  - Validates `from < to`, returns 400 on invalid range
- The backend is **date-range agnostic** — one endpoint, one query shape. The UI decides granularity:
  - Monthly: `?from=2026-02-01&to=2026-03-01`
  - Weekly: `?from=2026-02-09&to=2026-02-16`
  - Daily: `?from=2026-02-12&to=2026-02-13`
  - Custom: `?from=2026-01-15&to=2026-02-20`
- Wire summary service/handler into router

## Capabilities

### New Capabilities
- `spending-summary`: Date-range agnostic spending dashboard — total spent, per-category breakdown, per-participant breakdown with funded/balance computation, all via on-the-fly SQL aggregation (no materialized state). Single endpoint handles month/week/day/custom queries.

### Modified Capabilities
- `http-server`: Add summary route under `/api/v1/workspaces/{id}/summary`

## Impact

- **API:** 1 new endpoint (`GET /summary?from=&to=`) — the most important read endpoint in the application
- **Database:** No new migrations. Uses existing `transactions`, `entries`, `categories` tables via aggregate queries. Performance relies on the `(workspace_id, date)` composite index created in the transactions migration.
- **Performance:** On-the-fly aggregation. At household scale (hundreds-low thousands of transactions/month), PostgreSQL computes these in <10ms. If 100k+ transactions/month ever occur, add a materialized view — the application code stays identical.
- **Code:** ~4 new files (1 domain read model, 1 repo, 1 service, 1 handler)
