## 1. Frontend — Remove Multi-User UI

- [x] 1.1 Remove member management section from WorkspaceSettingsPage (member table, add/remove member dialog, related state)
- [x] 1.2 Remove `useAddMember` and `useRemoveMember` hooks from `frontend/src/api/workspaces.ts` (keep `useWorkspace` and `useWorkspaces`)
- [x] 1.3 Stop sending `participant_id` in CreateTransactionDialog — set entry `participant_id` to `null`
- [x] 1.4 Rename sidebar nav item from "Fundings" to "Budgets" and update route path from `/fundings` to `/budgets`
- [x] 1.5 Remove unused multi-user types from `types/index.ts` if no longer imported (`WorkspaceMember`, `MemberRole`, `ParticipantSpending`)

## 2. Database — Extend Fundings Table

- [x] 2.1 Create migration `000007_add_category_to_fundings.up.sql`: add nullable `category_id BIGINT REFERENCES categories(id) ON DELETE CASCADE`, drop old unique constraint, create new unique index on `(workspace_id, user_id, year_month, COALESCE(category_id, 0))`
- [x] 2.2 Create migration `000007_add_category_to_fundings.down.sql`: drop new index, remove `category_id` column, restore original unique constraint

## 3. Backend — Update Funding Model & Store

- [x] 3.1 Add `CategoryID *CategoryID` field to `model.Funding` (pointer for nullable)
- [x] 3.2 Update sqlc queries in `db/queries/fundings.sql`: add `category_id` to UpsertFunding params and ON CONFLICT clause, include `category_id` in SELECT results for ListFundingsByWorkspace and GetFundingsByWorkspaceAndUsers
- [x] 3.3 Add new sqlc query `DeleteFunding` (delete by id + workspace_id for safety)
- [x] 3.4 Add new sqlc query `GetOverallBudgetByWorkspace` and `GetCategoryBudgetsByWorkspace` to fetch category-level budgets for a date range (used by summary)
- [x] 3.5 Run `make sqlc` to regenerate and update `store/funding_store.go` mappers (`toModelFunding`) to handle nullable `category_id`

## 4. Backend — Update Funding Service & Handler

- [x] 4.1 Update `service.RecordFundingInput` to include optional `CategoryID *model.CategoryID`
- [x] 4.2 Update `service.FundingService.Record()` to pass `CategoryID` through to store
- [x] 4.3 Update `handler.recordFundingRequest` to include optional `category_id` field and make `user_id` optional (default to caller)
- [x] 4.4 Update `handler.RecordFunding` to default `user_id` to caller's ID when omitted
- [x] 4.5 Add `DeleteFunding` method to `FundingStore`, `FundingService`, and `FundingHandler` (DELETE /api/v1/workspaces/{id}/fundings/{fundingId})
- [x] 4.6 Register DELETE route in `handler/router.go`
- [x] 4.7 Update `fundingResponse()` to include `category_id` (nullable) in JSON output

## 5. Backend — Extend Summary with Budget Data

- [x] 5.1 Add `OverallBudget *Money` field to `model.SpendingSummary`
- [x] 5.2 Add `Budget *Money` field to `model.CategorySpending`
- [x] 5.3 Update `SummaryService.GetSummary()` to fetch overall budget (fundings where `category_id IS NULL`) and per-category budgets for the date range, summing across months
- [x] 5.4 Update `handler.SummaryHandler` response to include `overall_budget` and per-category `budget` fields

## 6. Frontend — Budgets Page

- [x] 6.1 Create `BudgetsPage.tsx` with month picker (default current month), overall budget input, and per-category budget inputs using workspace categories
- [x] 6.2 Add API hooks: `useBudgets(workspaceId, month)` to fetch fundings for a month, `useUpsertBudget()` to save, `useDeleteBudget()` to remove
- [x] 6.3 Implement save logic: upsert changed budget values (overall + per-category) on button click
- [x] 6.4 Replace FundingsPage import/route in App router with BudgetsPage at `/budgets`

## 7. Frontend — Dashboard Budget Widgets

- [x] 7.1 Update `Summary` type in `types/index.ts` to include `overall_budget` (nullable Money) and `budget` (nullable Money) on each category entry
- [x] 7.2 Create `BudgetProgressBar` component: progress bar with green/yellow/red color coding based on spent/budget ratio (<75% green, 75-100% yellow, >100% red)
- [x] 7.3 Update DashboardPage "Total Spent" card: when `overall_budget` is present, show spent/budget with progress bar; otherwise show plain amount as today
- [x] 7.4 Add category budget section to DashboardPage: horizontal bars for each category with a non-null budget, using `BudgetProgressBar`; hidden when no categories have budgets

## 8. Verification

- [x] 8.1 Run `make sqlc` and `go vet ./...` in backend — verify no compilation errors
- [x] 8.2 Run `npm run build && npm run lint` in frontend — verify clean build
- [x] 8.3 Run `make test` in backend — verify existing tests still pass
