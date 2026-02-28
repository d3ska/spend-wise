## 1. Model Entities

- [x] 1.1 Create `model/workspace.go` — `WorkspaceID` typed wrapper, `WorkspaceType` (TypePrivate/TypeShared), `Workspace` struct (ID, Name, Description, Type, OwnerID, CreatedAt, UpdatedAt), `Validate()` method returning sentinel errors
- [x] 1.2 Add `MemberRole` to `model/workspace.go` — `RoleOwner`/`RoleEditor`/`RoleViewer` constants, `Level()` method for comparison, `WorkspaceMember` struct (WorkspaceID, UserID, Role, CreatedAt), `CanEdit()`/`CanView()` methods
- [x] 1.3 Create `model/category.go` — `CategoryID` typed wrapper, `Category` struct (ID, WorkspaceID, Name, Icon, CreatedAt, UpdatedAt)

## 2. Database Migrations

- [x] 2.1 Create `migrations/000002_create_workspaces.up.sql` — `workspace_type` enum, `member_role` enum, `workspaces` table (BIGSERIAL PK, name, description, type, owner_id FK to users, timestamps), `workspace_members` table (composite PK workspace_id+user_id, role, created_at, ON DELETE CASCADE)
- [x] 2.2 Create `migrations/000002_create_workspaces.down.sql` — drop workspace_members, workspaces, enums
- [x] 2.3 Create `migrations/000003_create_categories.up.sql` — `categories` table (BIGSERIAL PK, workspace_id FK ON DELETE CASCADE, name, icon, timestamps), unique (workspace_id, name)
- [x] 2.4 Create `migrations/000003_create_categories.down.sql` — drop categories

## 3. Queries and Store

- [x] 3.1 Create `db/queries/workspaces.sql` — sqlc queries: InsertWorkspace, GetWorkspaceByID, ListWorkspacesByUser, UpdateWorkspace, DeleteWorkspace, InsertMember, GetMember, ListMembers, DeleteMember
- [x] 3.2 Create `db/queries/categories.sql` — sqlc queries: InsertCategory, GetCategoryByID, ListCategoriesByWorkspace, UpdateCategory, DeleteCategory
- [x] 3.3 Run `sqlc generate` to update generated code in `store/`
- [x] 3.4 Create `store/workspace_store.go` — `WorkspaceStore` with methods: Create, GetByID, ListByUser, Update, Delete, AddMember, GetMember, ListMembers, RemoveMember (with model mapping and sentinel error wrapping)
- [x] 3.5 Create `store/category_store.go` — `CategoryStore` with methods: Create, GetByID, ListByWorkspace, Update, Delete (with model mapping and sentinel error wrapping)

## 4. Service Layer

- [x] 4.1 Create `service/workspace_service.go` — `WorkspaceService` struct with workspace store, category store, user store dependencies. Co-located DTOs: CreateWorkspaceInput, UpdateWorkspaceInput, AddMemberInput, CreateCategoryInput, UpdateCategoryInput. Methods:
  - `requireMembership(ctx, wsID, userID, minRole)` — returns member or error
  - `CreateWorkspace` — validate, persist, auto-add creator as owner
  - `GetWorkspace`, `ListWorkspaces`, `UpdateWorkspace`, `DeleteWorkspace` — with permission checks
  - `AddMember`, `RemoveMember` — owner only
  - `CreateCategory`, `ListCategories`, `UpdateCategory`, `DeleteCategory` — editor+ for writes, viewer+ for reads

## 5. HTTP Handlers

- [x] 5.1 Create `handler/handler_workspace.go` — `WorkspaceHandler` struct wrapping service. Methods: CreateWorkspace, ListWorkspaces, GetWorkspace, UpdateWorkspace, DeleteWorkspace, AddMember, RemoveMember. Parse workspace ID from URL param, UserID from context.
- [x] 5.2 Create `handler/handler_category.go` — `CategoryHandler` struct wrapping service. Methods: CreateCategory, ListCategories, UpdateCategory, DeleteCategory. Parse workspace ID and category ID from URL params.

## 6. Router Wiring

- [x] 6.1 Update `handler/router.go` — add `WorkspaceHandler` and `CategoryHandler` to `RouterConfig`, register routes under the protected group:
  - `POST /api/v1/workspaces`, `GET /api/v1/workspaces`, `GET /api/v1/workspaces/{id}`, `PUT /api/v1/workspaces/{id}`, `DELETE /api/v1/workspaces/{id}`
  - `POST /api/v1/workspaces/{id}/members`, `DELETE /api/v1/workspaces/{id}/members/{userID}`
  - `POST /api/v1/workspaces/{id}/categories`, `GET /api/v1/workspaces/{id}/categories`, `PUT /api/v1/workspaces/{id}/categories/{catID}`, `DELETE /api/v1/workspaces/{id}/categories/{catID}`
- [x] 6.2 Update `cmd/api/main.go` — create WorkspaceStore, CategoryStore, WorkspaceService, WorkspaceHandler, CategoryHandler, pass to NewRouter

## 7. Verification

- [x] 7.1 Run `go build ./...` — must compile with zero errors
- [x] 7.2 Run `go vet ./...` — must pass with zero warnings
- [x] 7.3 Run `go test -count=1 ./...` — all tests must pass
