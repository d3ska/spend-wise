## Context

SpendWise is a collaborative household financial ledger. Multiple users share workspaces and mutate shared data (transactions, categories, fundings, rules, members). The frontend uses React Query for data fetching with stale-while-revalidate caching. Currently, changes made by one user are invisible to other connected users until a refetch is triggered (manual refresh or stale timer). The backend is a monolithic Go API using chi router with JWT cookie auth.

## Goals / Non-Goals

**Goals:**
- Deliver instant (~1s) UI updates to all connected workspace members when any workspace-scoped mutation occurs
- Zero database changes — purely in-memory notification layer
- Minimal handler-level integration (one-liner broadcast calls)
- Leverage browser-native EventSource API with automatic reconnection

**Non-Goals:**
- Payload delivery (events carry only the event type, not the data — clients refetch via existing API)
- Cross-server scalability (in-memory hub is single-process; Redis pub/sub can be added later if needed)
- User-scoped events (bank connections are user-scoped, not workspace-scoped — skip for now)
- Offline/missed-event recovery (clients reconnect and refetch naturally via React Query)

## Decisions

### 1. SSE over WebSockets

**Decision:** Use Server-Sent Events (SSE), not WebSockets.

**Rationale:** Communication is unidirectional (server → client). SSE uses plain HTTP, works through proxies, and the browser `EventSource` API handles reconnection automatically. WebSockets would add bidirectional overhead with no benefit here.

**Alternatives considered:**
- *WebSockets*: More complex (upgrade handshake, ping/pong), no advantage for one-way push
- *Long polling*: Higher latency and server load per update cycle

### 2. In-memory pub/sub hub

**Decision:** A single `SSEHub` struct with `sync.RWMutex` guarding a `map[int64]map[chan string]struct{}` (workspaceID → set of client channels).

**Rationale:** Simple, zero-dependency, fits the single-process deployment model. The hub lives for the lifetime of the process. Client count is bounded by concurrent browser tabs (household app — dozens at most, not thousands).

**Alternatives considered:**
- *Redis pub/sub*: Unnecessary for a single-server household app; can migrate later if needed
- *NATS/Kafka*: Extreme overkill for this scale

### 3. Event types without payloads

**Decision:** Broadcast only the event type string (e.g., `transaction_changed`). Clients invalidate the relevant React Query cache keys and refetch through existing API endpoints.

**Rationale:** Keeps the SSE layer trivial — no serialization, no partial-update logic, no versioning. React Query already handles efficient refetching and deduplication.

**Alternatives considered:**
- *Full payloads*: Would require serializing domain objects in handlers, versioning event schemas, and handling partial updates — significant complexity for marginal latency improvement

### 4. Hub injected via handler struct field

**Decision:** Add a `hub *SSEHub` field to each handler struct that has workspace-scoped mutations (WorkspaceHandler, InviteHandler, CategoryHandler, TransactionHandler, FundingHandler, RuleHandler). Constructor functions gain an optional hub parameter.

**Rationale:** Follows the existing dependency injection pattern (handlers already receive service pointers). Keeps broadcast calls co-located with the mutation response.

### 5. Six coarse-grained event types

**Decision:** Use 6 event types covering all 21+ workspace-scoped mutations:

| Event Type | Triggers |
|---|---|
| `workspace_changed` | Create/Update/Delete workspace |
| `member_changed` | Add/Remove member, AcceptInvite, UpdateMemberRole |
| `transaction_changed` | Create/Update/Delete transaction, Import, ApplyRules |
| `category_changed` | Create/Update/Delete category |
| `funding_changed` | Record/Delete funding |
| `rule_changed` | Create/Update/Delete/Toggle rule |

**Rationale:** Coarse events are simpler to maintain. React Query invalidation by key prefix naturally refetches only the relevant data. Fine-grained events can be introduced later without breaking clients.

### 6. Heartbeat keepalive

**Decision:** Send an SSE comment (`:heartbeat\n\n`) every 30 seconds.

**Rationale:** Prevents intermediate proxies and load balancers from closing idle connections. SSE comment lines are ignored by the EventSource API.

## Risks / Trade-offs

- **[Single-process limitation]** In-memory hub does not work across multiple backend instances. → Acceptable for a household app. Mitigation: migrate to Redis pub/sub if horizontal scaling is ever needed.
- **[Thundering herd on reconnect]** If the server restarts, all clients reconnect simultaneously and refetch all data. → At household scale (2-10 users) this is negligible. EventSource has built-in jitter.
- **[No auth on reconnect]** EventSource doesn't send cookies on some older browsers. → Modern browsers (Chrome, Firefox, Safari) support `withCredentials: true` on EventSource. The endpoint is behind JWT middleware.
- **[WriteTimeout conflict]** Go's `http.Server.WriteTimeout` will kill long-lived SSE connections. → The SSE handler must use `http.Hijack` or the response controller to disable/extend the write deadline per-connection.
