## Why

SpendWise is a collaborative household financial ledger where multiple users share workspaces. Currently, when one user creates or modifies a transaction, category, funding, or rule, other connected workspace members only see the change after a manual page refresh or when React Query's stale timer triggers a refetch. This creates a disjointed experience for shared workspaces. Server-Sent Events (SSE) will give all connected members instant UI updates when any workspace-scoped mutation occurs.

## What Changes

- Add an in-memory pub/sub hub on the backend that tracks SSE client connections per workspace
- Add an SSE streaming endpoint (`GET /api/v1/workspaces/{id}/events`) behind JWT auth
- Broadcast lightweight event-type messages from every workspace-scoped mutation handler (transactions, categories, fundings, rules, members, workspace settings)
- Add a frontend `useSSE` hook that opens an `EventSource` connection per workspace and invalidates the relevant React Query caches when events arrive
- Wire the hook into `AppShell` so it covers all pages

## Capabilities

### New Capabilities
- `sse-streaming`: Server-Sent Events hub, endpoint, and broadcast integration for real-time workspace notifications

### Modified Capabilities

(none - no existing spec-level requirements change; this is purely additive)

## Impact

- **Backend handlers**: 6 handler files gain a `hub` field and one-liner broadcast calls after successful mutations (`handler_workspace.go`, `handler_invite.go`, `handler_category.go`, `handler_transaction.go`, `handler_funding.go`, `handler_rule.go`)
- **Backend wiring**: `router.go` and `cmd/api/main.go` create and pass the SSE hub
- **Frontend**: New `useSSE` hook, wired once in `AppShell.tsx`
- **No database changes**
- **No new dependencies** (SSE uses stdlib `net/http`; frontend uses native `EventSource` API)
