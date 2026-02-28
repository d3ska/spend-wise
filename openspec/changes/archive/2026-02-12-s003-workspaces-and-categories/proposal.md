## Why

Every financial entity in SpendWise belongs to a Workspace — it is the top-level organizational unit. Workspaces can be private (single user) or shared (multiple users with role-based permissions). Categories group transactions for spending breakdowns. Both must exist before transactions can be created, making this the next incremental step after authentication.

**Depends on:** `s002-auth-and-users` (User entity, JWT middleware, user store)

## What Changes

- Create `Workspace` entity in `model/workspace.go` with typed `WorkspaceID`, `WorkspaceType` (private|shared), mandatory name and description, `Validate()` method
- Create `WorkspaceMember` entity with `MemberRole` (owner|editor|viewer), `CanEdit()` / `CanView()` methods
- Create `Category` entity in `model/category.go` with typed `CategoryID`, workspace-scoped with unique (workspace_id, name)
- Create database migration `000002_create_workspaces` with `workspace_type` and `member_role` PostgreSQL enums, composite PK on workspace_members
- Create database migration `000003_create_categories` with unique (workspace_id, name) constraint
- Create sqlc queries (`db/queries/workspaces.sql`, `db/queries/categories.sql`)
- Run `sqlc generate` to update `store/` generated code
- Create `store/workspace_store.go` and `store/category_store.go`
- Create `service/workspace_service.go` with co-located DTOs:
  - `Create` — validates workspace, persists, auto-adds creator as owner member
  - `List` — returns workspaces the user is a member of
  - `Get`, `Update`, `Delete` — with permission checks
  - `AddMember`, `RemoveMember` — owner manages members
  - `requireMembership(ctx, wsID, userID, minRole)` — shared permission helper
- Create `handler/handler_workspace.go` with workspace + member endpoints
- Create `handler/handler_category.go` with category CRUD endpoints (nested under workspace)
- Wire workspace and category services/handlers into router

## Capabilities

### New Capabilities
- `workspace-management`: Workspace CRUD (private/shared), mandatory description, role-based member management (owner/editor/viewer), permission enforcement via `requireMembership` helper
- `category-management`: Category CRUD scoped to workspace, unique name constraint per workspace, icon support

### Modified Capabilities
- `http-server`: Add workspace and category routes to the Chi router under `/api/v1/workspaces`

## Impact

- **API:** 11 new endpoints — workspace CRUD (5), member management (2), category CRUD (4)
- **Database:** Migrations 002-003 create `workspaces`, `workspace_members`, `categories` tables with PG enums
- **Authorization model:** Establishes the permission system (owner/editor/viewer) used by all subsequent changes
- **Code:** ~8 new files (2 model, 2 store, 2 service, 2 handler) + migration and query files
