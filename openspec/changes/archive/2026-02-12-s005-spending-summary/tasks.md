## 1. Model Read Models

- [x] 1.1 Create `model/summary.go` — `SpendingSummary` struct (WorkspaceID, From, To time.Time, TotalSpent Money, ByCategory []CategorySpending, ByParticipant []ParticipantSpending), `CategorySpending` struct (CategoryID, Name, Icon string, Spent Money), `ParticipantSpending` struct (UserID, DisplayName string, Spent Money, Funded Money, Balance Money)

## 2. Summary Store

- [x] 2.1 Create `store/summary_store.go` — `SummaryStore` with `*pgxpool.Pool` dependency. Three methods using manual `pool.Query` (not sqlc): `TotalSpent(ctx, wsID, from, to) → Money`, `SpentByCategory(ctx, wsID, from, to) → []CategorySpending`, `SpentByParticipant(ctx, wsID, from, to) → []ParticipantSpending`. All queries use `WHERE t.date >= $from AND t.date < $to` with the existing `(workspace_id, date)` index.

## 3. Service Layer

- [x] 3.1 Create `service/summary_service.go` — `SummaryService` struct with summary store, workspace store dependencies. Method: `GetSummary(ctx, userID, wsID, from, to) → SpendingSummary` — verify viewer+ permission via workspace store, run all three store queries, assemble SpendingSummary with Funded/Balance stubbed to zero until funding feature lands.

## 4. HTTP Handler

- [x] 4.1 Create `handler/handler_summary.go` — `SummaryHandler` struct wrapping service. Method: `GetSummary(w, r)` — parse workspace ID from URL, parse `from`/`to` from query params (YYYY-MM-DD), call service, return JSON with money as `{"amount": "string", "currency": "string"}`.

## 5. Router Wiring

- [x] 5.1 Update `handler/router.go` — add `SummaryHandler` to `RouterConfig`, register `GET /api/v1/workspaces/{id}/summary` under protected workspace routes
- [x] 5.2 Update `cmd/api/main.go` — create SummaryStore, SummaryService, SummaryHandler, pass to NewRouter

## 6. Verification

- [x] 6.1 Run `go build ./...` — must compile with zero errors
- [x] 6.2 Run `go vet ./...` — must pass with zero warnings
- [x] 6.3 Run `go test -count=1 ./...` — all tests must pass
