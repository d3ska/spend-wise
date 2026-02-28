## ADDED Requirements

### Requirement: SSE endpoint streams workspace events to authenticated members
The system SHALL expose an SSE endpoint at `GET /api/v1/workspaces/{id}/events` that streams real-time event notifications to authenticated workspace members. The endpoint MUST require a valid JWT and MUST verify the caller is a member of the workspace.

#### Scenario: Authenticated member connects to SSE endpoint
- **WHEN** an authenticated workspace member sends a GET request to `/api/v1/workspaces/{id}/events`
- **THEN** the server responds with `Content-Type: text/event-stream`, `Cache-Control: no-cache`, and `Connection: keep-alive` headers and holds the connection open

#### Scenario: Unauthenticated user is rejected
- **WHEN** a request without a valid JWT is sent to `/api/v1/workspaces/{id}/events`
- **THEN** the server responds with 401 Unauthorized

#### Scenario: Non-member user is rejected
- **WHEN** an authenticated user who is not a member of the workspace connects to the SSE endpoint
- **THEN** the server responds with 403 Forbidden

### Requirement: Workspace mutations broadcast event-type notifications
The system SHALL broadcast an event-type string to all connected SSE clients for a workspace whenever a workspace-scoped mutation succeeds. The event data SHALL contain only the event type, not the mutated payload. The following event types SHALL be supported:

- `workspace_changed` — workspace created, updated, or deleted
- `member_changed` — member added, removed, invite accepted, or role updated
- `transaction_changed` — transaction created, updated, deleted, imported, or rules applied
- `category_changed` — category created, updated, or deleted
- `funding_changed` — funding recorded or deleted
- `rule_changed` — rule created, updated, deleted, or toggled

#### Scenario: Transaction created triggers event
- **WHEN** a user creates a transaction in workspace 10
- **THEN** all SSE clients connected to workspace 10 receive an event with type `transaction_changed`

#### Scenario: Category deleted triggers event
- **WHEN** a user deletes a category in workspace 10
- **THEN** all SSE clients connected to workspace 10 receive an event with type `category_changed`

#### Scenario: Member added triggers event
- **WHEN** a user accepts an invite to workspace 10
- **THEN** all SSE clients connected to workspace 10 receive an event with type `member_changed`

### Requirement: SSE connection sends periodic heartbeat
The system SHALL send an SSE comment line (`:heartbeat`) at least every 30 seconds on each open connection to prevent proxy/load-balancer timeouts.

#### Scenario: Idle connection receives heartbeat
- **WHEN** an SSE connection has been open for 30 seconds without any event
- **THEN** the server sends a `:heartbeat` comment line to keep the connection alive

### Requirement: SSE hub manages client lifecycle
The system SHALL maintain an in-memory hub that tracks connected SSE clients per workspace. When a client disconnects (browser closed, network lost), the hub MUST remove the client channel and free resources.

#### Scenario: Client disconnects cleanly
- **WHEN** a connected SSE client closes their browser tab
- **THEN** the hub removes the client's channel from the workspace subscriber set

#### Scenario: Multiple clients on same workspace
- **WHEN** two clients are connected to the same workspace and a transaction is created
- **THEN** both clients receive the `transaction_changed` event

### Requirement: Frontend invalidates queries on SSE events
The frontend SHALL open an EventSource connection to the SSE endpoint for the active workspace and invalidate the appropriate React Query cache keys when events are received. The mapping SHALL be:

- `workspace_changed` → `["workspaces"]`
- `member_changed` → `["workspaces", wsID, "members"]`, `["workspaces"]`
- `transaction_changed` → `["transactions", wsID]`, `["summary", wsID]`
- `category_changed` → `["categories", wsID]`
- `funding_changed` → `["fundings", wsID]`, `["summary", wsID]`
- `rule_changed` → `["rules", wsID]`

#### Scenario: Transaction changed event triggers refetch
- **WHEN** the frontend receives a `transaction_changed` SSE event for workspace 5
- **THEN** React Query invalidates `["transactions", 5]` and `["summary", 5]` causing automatic refetch of visible data

#### Scenario: Workspace switch closes old connection
- **WHEN** the user switches from workspace 5 to workspace 8
- **THEN** the EventSource for workspace 5 is closed and a new one is opened for workspace 8
