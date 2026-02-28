## Context

Authentication and user management (`s002-auth-and-users`) are complete. The project has the flat package layout (`model/`, `store/`, `handler/`, `service/`, `auth/`, `config/`), pgx connection pool, Chi router with JWT middleware, and sqlc for type-safe queries.

This change introduces Workspaces (the top-level container for all financial data) and Categories (for grouping transactions). Every subsequent feature (transactions, funding, summaries, rules) is scoped to a workspace and requires the permission model established here.

## Goals / Non-Goals

**Goals:**
- Workspace CRUD with private/shared types
- Role-based access control (owner/editor/viewer) via workspace members
- `requireMembership` permission helper reusable by all workspace-scoped services
- Category CRUD scoped to workspace with unique name constraint
- Auto-add creator as owner on workspace creation

**Non-Goals:**
- No workspace invitations via email notification — just the API to add members by user ID
- No workspace transfer (change owner) — can be added later
- No category nesting/hierarchy — flat list per workspace
- No default categories on workspace creation

## Decisions

### 1. Permission check as a shared helper, not middleware

`requireMembership(ctx, wsID, userID, minRole)` is a function in the service layer, not a Chi middleware. Reason: the workspace ID comes from different places (URL param for workspace endpoints, request body for nested resources). A middleware would need to know the extraction strategy for every route.

**Alternative considered:** Chi middleware that reads `{id}` from URL. Breaks for nested resources where workspace ID comes from a parent lookup, not the URL directly.

### 2. MemberRole as ordered enum for permission comparison

`MemberRole` has a `Level()` method returning an int: viewer=1, editor=2, owner=3. Permission checks use `role.Level() >= minRole.Level()`. This avoids switch/case chains.

**Alternative considered:** Explicit `CanEdit()`, `CanDelete()` per role. More methods to maintain; the level-based approach is simpler and extensible.

### 3. Workspace deletion is soft — only by owner

Only the owner can delete a workspace. Deletion cascades to members, categories, and (in future changes) transactions. The database uses `ON DELETE CASCADE` for simplicity.

**Alternative considered:** Soft delete with `deleted_at` column. Adds complexity (all queries need `WHERE deleted_at IS NULL`) without clear value at this stage.

### 4. Category service is thin — no separate service file

Category operations are simple CRUD scoped to a workspace. They live in `service/workspace_service.go` alongside workspace operations rather than a separate file, since every category operation starts with a `requireMembership` check. If the file grows too large, category methods can be extracted later.

**Alternative considered:** Separate `service/category_service.go`. Creates two files that both need the same permission infrastructure. Merge when small, split when large.

### 5. Workspace type is informational, not enforced

`WorkspaceType` (private/shared) is a label for the UI. The backend doesn't enforce "private means single member" — a private workspace can technically have members added. The UI uses the type to decide whether to show sharing controls.

**Alternative considered:** Enforce private = exactly 1 member. Adds validation complexity for little gain; the owner controls who gets added regardless.

## Risks / Trade-offs

- **[Risk] CASCADE deletion** → Deleting a workspace removes all related data. Acceptable for an explicit owner action. Could add a confirmation flag in the API later.
- **[Trade-off] No separate category service** → workspace_service.go will be larger. Acceptable while the service has < 300 lines; split at that point.
- **[Trade-off] Members by user ID, not email** → Frontend must look up user IDs. Simpler backend; a "search user by email" endpoint can be added if needed.
