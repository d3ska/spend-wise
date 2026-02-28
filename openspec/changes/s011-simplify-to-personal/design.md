## Context

SpendWise has a full shared-workspace layer (members, roles, participant tracking) that is unused in practice. The `fundings` table currently tracks per-user monthly contributions with a unique constraint on `(workspace_id, user_id, year_month)`. The summary endpoint already merges funding data into participant spending. The frontend dashboard shows total spent and category breakdown but has no budget visualization.

The goal is to strip multi-user UI, repurpose fundings as personal budget targets, and surface budget-vs-actual on the dashboard.

## Goals / Non-Goals

**Goals:**

- Remove all multi-user UI from the frontend (members, participant assignment, workspace type)
- Extend fundings to support per-category budgets alongside the existing overall budget
- Return budget data from the summary endpoint so the dashboard can show budget vs actual
- Build a clean "Budgets" page replacing the raw FundingsPage
- Add budget progress widgets to the dashboard

**Non-Goals:**

- Dropping backend database tables or API fields (keep for future re-enablement)
- Removing backend permission checks or member logic
- Multi-currency budget support (budgets use workspace default currency, PLN)
- Budget rollover between months
- Budget alerts/notifications

## Decisions

### 1. Extend `fundings` table with nullable `category_id`

**Decision:** Add a nullable `category_id BIGINT REFERENCES categories(id)` column to the `fundings` table. A row with `category_id = NULL` represents the overall monthly budget. A row with a `category_id` value represents a per-category budget for that month.

**Rationale:** Reuses the existing table and upsert logic rather than creating a separate `budgets` table. The existing `(workspace_id, user_id, year_month)` unique constraint becomes `(workspace_id, user_id, year_month, COALESCE(category_id, 0))` to allow both overall and per-category entries for the same month.

**Alternative considered:** Separate `budgets` table — rejected because it would duplicate the same structure (workspace, user, month, amount) and require parallel CRUD logic.

### 2. Use a single migration to add column + update constraint

**Decision:** One new migration (`000008_add_category_to_fundings`) that:
1. Adds `category_id BIGINT REFERENCES categories(id) ON DELETE CASCADE` (nullable)
2. Drops the old unique constraint
3. Creates a new unique index on `(workspace_id, user_id, year_month, COALESCE(category_id, 0))`

Using `COALESCE(category_id, 0)` in the unique index handles NULL correctly (NULLs are distinct in PostgreSQL unique constraints, so without COALESCE, multiple rows with `category_id = NULL` for the same month would be allowed).

### 3. Extend summary response with budget data

**Decision:** Add `overall_budget` (Money | null) and a `budget` field (Money | null) to each `CategorySpending` entry in the summary response. The backend queries fundings for the date range and attaches them.

**Rationale:** The summary endpoint already fetches fundings for the date range (to calculate participant funded amounts). We extend this to also pull category-level budgets and the overall budget, returning them inline. No new endpoint needed.

**Response shape changes:**
```json
{
  "total_spent": { "amount": "1234.00", "currency": "PLN" },
  "overall_budget": { "amount": "5000.00", "currency": "PLN" },
  "by_category": [
    {
      "category_id": 1,
      "name": "Groceries",
      "icon": "🛒",
      "spent": { "amount": "400.00", "currency": "PLN" },
      "budget": { "amount": "600.00", "currency": "PLN" }
    }
  ],
  "by_participant": [...]
}
```

### 4. Implicit user_id on funding/budget API calls

**Decision:** When recording a budget, the handler uses the caller's user ID from the JWT token instead of requiring `user_id` in the request body. The `user_id` field becomes optional (ignored if provided, defaults to caller). This simplifies the frontend — no need to know or send a user ID.

**Rationale:** In personal mode, the only user setting budgets is the logged-in user. Keeps the API backwards-compatible (old requests with `user_id` still work, but it's no longer required).

### 5. Frontend budget management UX

**Decision:** Replace `FundingsPage` with a `BudgetsPage` that shows:
- A month picker (defaults to current month)
- An "Overall Budget" input at the top
- A table/list of categories with a budget input field next to each
- Save button that upserts all changed values

**Rationale:** Batch editing is more natural than individual dialogs for budget setting. User sees all categories at once and sets targets for the selected month.

### 6. Dashboard budget widgets

**Decision:** When `overall_budget` is present in the summary response:
- Replace the plain "Total Spent" card with a progress bar showing spent/budget
- Add a percentage indicator and color coding (green < 75%, yellow 75-100%, red > 100%)

When category-level `budget` values are present:
- Add a horizontal bar chart or list below the pie chart showing each category's spent vs budget
- Same color coding logic

When no budgets are defined, the dashboard looks exactly the same as today.

### 7. Frontend simplification (multi-user removal)

**Decision:** Remove from frontend only — no backend changes:
- `WorkspaceSettingsPage`: remove member management section (keep name/description editing)
- `CreateTransactionDialog`: stop sending `participant_id` in entries (send `null`)
- `FundingsPage`: replaced entirely by `BudgetsPage`
- `types/index.ts`: keep types but stop using `WorkspaceMember`, `ParticipantSpending` in UI
- Sidebar/nav: rename "Fundings" to "Budgets"

## Risks / Trade-offs

**[Migration on existing data]** → Existing funding rows will have `category_id = NULL`, which correctly represents overall budgets. No data migration needed — the new column is nullable and existing rows remain valid.

**[COALESCE in unique index]** → Slightly non-obvious constraint. If category 0 ever exists, it would conflict with NULL entries. Mitigation: category IDs are auto-increment from 1, so 0 is never a valid ID.

**[Summary endpoint does more work]** → Now fetches both participant fundings and category budgets. Mitigation: it's one extra WHERE clause on the same table; negligible performance impact.

**[Frontend still sends participant_id=null]** → Backend accepts null participant_id on entries already (column is nullable). No backend change needed, just stop populating it from frontend.
