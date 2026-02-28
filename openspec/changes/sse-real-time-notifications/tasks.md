## 1. SSE Hub (in-memory pub/sub)

- [x] 1.1 Create `handler/sse_hub.go` with `SSEHub` struct (`sync.RWMutex`, `map[int64]map[chan string]struct{}`), `NewSSEHub()`, `Subscribe(workspaceID) (chan, unsubscribe)`, `Broadcast(workspaceID, eventType)`

## 2. SSE Endpoint

- [x] 2.1 Create `handler/handler_sse.go` with SSE handler: `GET /api/v1/workspaces/{id}/events` — set SSE headers, verify workspace membership, subscribe to hub, stream events, send 30s heartbeat, unsubscribe on disconnect, disable write deadline
- [x] 2.2 Register SSE route in `handler/router.go` — add `SSEHub` to `RouterConfig`, register `GET /workspaces/{id}/events` inside the protected workspace group
- [x] 2.3 Create `SSEHub` in `cmd/api/main.go` and pass it to `RouterConfig`

## 3. Broadcast from Mutation Handlers

- [x] 3.1 Add `hub *SSEHub` field to `WorkspaceHandler`, update constructor, broadcast `workspace_changed` after CreateWorkspace/UpdateWorkspace/DeleteWorkspace and `member_changed` after AddMember/RemoveMember
- [x] 3.2 Add `hub *SSEHub` field to `InviteHandler`, update constructor, broadcast `member_changed` after AcceptInvite and UpdateMemberRole
- [x] 3.3 Add `hub *SSEHub` field to `CategoryHandler`, update constructor, broadcast `category_changed` after CreateCategory/UpdateCategory/DeleteCategory
- [x] 3.4 Add `hub *SSEHub` field to `TransactionHandler`, update constructor, broadcast `transaction_changed` after CreateTransaction/UpdateTransaction/DeleteTransaction/ImportTransactions/ApplyRules
- [x] 3.5 Add `hub *SSEHub` field to `FundingHandler`, update constructor, broadcast `funding_changed` after RecordFunding/DeleteFunding
- [x] 3.6 Add `hub *SSEHub` field to `RuleHandler`, update constructor, broadcast `rule_changed` after CreateRule/UpdateRule/DeleteRule/ToggleRule
- [x] 3.7 Update handler construction calls in `cmd/api/main.go` to pass `sseHub`

## 4. Frontend SSE Hook

- [x] 4.1 Create `frontend/src/hooks/useSSE.ts` — open `EventSource` to `/api/v1/workspaces/${workspaceId}/events` with credentials, map event types to React Query key invalidations, clean up on unmount or workspace change
- [x] 4.2 Wire `useSSE(workspaceId)` into `AppShellInner` in `frontend/src/components/layout/AppShell.tsx`

## 5. Verification

- [x] 5.1 Run `go vet ./...` and `go test -count=1 ./...` — all pass
- [x] 5.2 Run `npx tsc --noEmit` in frontend — compiles cleanly
- [ ] 5.3 Manual test: open two browser tabs on the same workspace, create a transaction in one, verify the other tab updates within ~1 second without refresh
