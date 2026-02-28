## Why

SpendWise currently has shared-workspace scaffolding (members, roles, participant tracking, split entries) but none of it is battle-tested or has meaningful UX. This complexity slows development and creates confusing UI. Simplify to a personal finance tracker first — one user, one workspace — and repurpose fundings as personal budgets (overall + per-category) with dashboard visualization. Multi-user features can be re-added later once the core experience is solid.

## What Changes

### Removals (frontend only — keep backend tables/APIs intact)

- **Remove member management UI** — drop the members table and add/remove member controls from WorkspaceSettingsPage
- **Remove participant references in transactions** — stop sending `participant_id` on entries; the backend column stays nullable
- **Remove participant spending from dashboard** — the `ByParticipant` summary data is fetched but never displayed; stop requesting it
- **Remove FundingsPage user_id input** — fundings become the current user's budget, no need to specify a user ID
- **Strip workspace type concept** — always create "personal" workspaces, hide type selector

### Modifications

- **Repurpose fundings as budgets** — a funding record becomes a monthly budget target:
  - **Overall budget**: funding with no category (existing behavior, `category_id = NULL`)
  - **Per-category budget**: funding tied to a specific category (new: add `category_id` column to `fundings` table)
  - Both are optional — the app works fine without any budgets defined
- **Backend: extend fundings table** — add nullable `category_id` foreign key; update store/service/handler to support it; adjust unique constraint to `(workspace_id, user_id, year_month, category_id)`
- **Backend: extend summary endpoint** — return budget data alongside spending (overall budget + per-category budgets for the queried range)
- **Frontend: new budget management UI** — replace raw FundingsPage with a cleaner "Budgets" page where user sets overall and per-category monthly amounts
- **Frontend: dashboard budget widgets** — when budgets are defined, show:
  - Overall: spent vs budget progress bar
  - Per-category: spent vs budget bars in the category breakdown
  - Visual indicator when over budget (red) vs on track (green/neutral)

## Capabilities

### New Capabilities

- `personal-budgets`: Monthly budget targets — overall and per-category — stored as extended fundings, with CRUD API and management UI
- `dashboard-budget-view`: Dashboard widgets showing budget vs actual spending with progress bars and over-budget indicators

### Modified Capabilities

_None — backend APIs remain backwards-compatible; we're adding fields, not changing existing behavior._

## Impact

- **Database**: migration to add `category_id` to `fundings` table + update unique constraint
- **Backend**: `model/funding.go`, `store/funding_store.go`, `service/funding_service.go`, `handler/handler_funding.go` — extend for category-level budgets; `model/summary.go`, `store/summary_store.go`, `service/summary_service.go` — include budget data in summary response
- **Frontend**: `FundingsPage` → rewrite as `BudgetsPage`; `DashboardPage` → add budget widgets; `WorkspaceSettingsPage` → remove member management; `CreateTransactionDialog` → stop sending participant_id; `types/index.ts` → update types
- **No breaking API changes** — existing endpoints get optional new fields
