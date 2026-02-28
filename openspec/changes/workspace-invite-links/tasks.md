## 1. Frontend Cleanup (WorkspaceType removal)

- [x] 1.1 Remove `type` field from `Workspace` interface in `frontend/src/types/index.ts`
- [x] 1.2 Remove `type` param from `useCreateWorkspace` mutation in `frontend/src/api/workspaces.ts`
- [x] 1.3 Remove `type: "shared"` from `CreateWorkspacePrompt.tsx` create call

## 2. Database & Queries (Invites)

- [x] 2.1 Create migration `000013_create_workspace_invites.up.sql` and `.down.sql` with `workspace_invites` table schema
- [x] 2.2 Add sqlc queries: `InsertInvite`, `GetInviteByCode` (JOIN workspaces + users for preview data), `MarkInviteUsed`
- [x] 2.3 Run `make sqlc` to regenerate store code

## 3. Database & Queries (Enriched Members + Member Count)

- [x] 3.1 Add sqlc query `ListMembersWithProfiles` — JOIN `workspace_members` with `users` to return display_name, email, avatar_url
- [x] 3.2 Add sqlc query `UpdateMemberRole` — update role for a specific workspace member
- [x] 3.3 Modify `ListWorkspacesByUser` to include a subquery count for member_count
- [x] 3.4 Run `make sqlc` to regenerate store code

## 4. Backend Model (Invite errors)

- [x] 4.1 Add sentinel errors to `model/errors.go`: `ErrInviteNotFound`, `ErrInviteExpired`, `ErrInviteUsed`, `ErrAlreadyMember`, `ErrCannotChangeOwnRole`, `ErrInvalidInviteRole`

## 5. Backend Store (Invite store)

- [x] 5.1 Create `store/invite_store.go` with `InviteStore` wrapping sqlc queries: `Create`, `GetByCode`, `MarkUsed`
- [x] 5.2 Map `pgx.ErrNoRows` to `model.ErrInviteNotFound` in `GetByCode`

## 6. Backend Store (Workspace store additions)

- [x] 6.1 Add `ListMembersWithProfiles` method to `WorkspaceStore` returning enriched member data
- [x] 6.2 Add `UpdateMemberRole` method to `WorkspaceStore`
- [x] 6.3 Updated `ListByUser` to map `ListWorkspacesByUserRow` including `MemberCount`
- [x] 6.4 Update `WorkspaceStoreIface` in `workspace_service.go` with new methods

## 7. Backend Service (Invite service)

- [x] 7.1 Create `service/invite_service.go` with `InviteService` struct and store interface
- [x] 7.2 Implement `CreateInvite(ctx, userID, wsID, role, expiresInHours)` — generate code, validate owner role, validate role is editor/viewer, persist
- [x] 7.3 Implement `PreviewInvite(ctx, code)` — return workspace name, inviter name, role, expiry; check expired/used status
- [x] 7.4 Implement `AcceptInvite(ctx, userID, code)` — validate not expired/used, check not already member, add member, mark used
- [x] 7.5 Write tests for `InviteService` (table-driven: valid creation, non-owner rejected, expired, used, already member, invalid role)

## 8. Backend Service (Workspace service additions)

- [x] 8.1 Add `ListMembersWithProfiles(ctx, userID, wsID)` to `WorkspaceService` with viewer+ permission check
- [x] 8.2 Add `UpdateMemberRole(ctx, userID, wsID, targetUserID, newRole)` to `WorkspaceService` — owner only, cannot change self, cannot set owner
- [x] 8.3 `ListWorkspaces` already returns member counts (store's `ListByUser` includes `MemberCount`)
- [x] 8.4 Write tests for new workspace service methods

## 9. Backend Handler (Invite endpoints)

- [x] 9.1 Create `handler/handler_invite.go` with `InviteHandler` struct
- [x] 9.2 Implement `CreateInvite` handler — `POST /workspaces/{id}/invites`, parse body `{ role, expires_in_hours }`, return invite with URL
- [x] 9.3 Implement `PreviewInvite` handler — `GET /invites/{code}`, public endpoint, return preview data
- [x] 9.4 Implement `AcceptInvite` handler — `POST /invites/{code}/accept`, requires auth, return workspace
- [x] 9.5 Register invite routes in `router.go` (preview + accept are top-level under `/api/v1/invites`, create is nested under workspace)
- [x] 9.6 Add new error mappings in `handleInviteError` for invite errors (410 for expired/used, 409 for already member)
- [ ] 9.7 Write handler tests for invite endpoints

## 10. Backend Handler (Member management endpoints)

- [x] 10.1 Implement `ListMembers` handler — `GET /workspaces/{id}/members`, return enriched member list
- [x] 10.2 Implement `UpdateMemberRole` handler — `PUT /workspaces/{id}/members/{userID}`, parse body `{ role }`
- [x] 10.3 Register new routes in `router.go` (GET members, PUT member role)
- [x] 10.4 Update workspace response to include `member_count` field
- [ ] 10.5 Write handler tests for member management endpoints

## 11. OpenAPI Spec Update

- [ ] 11.1 Add invite endpoints to `api/openapi.yaml` (POST create, GET preview, POST accept)
- [ ] 11.2 Add member management endpoints (GET list members, PUT update role)
- [ ] 11.3 Add `member_count` to `WorkspaceResponse` schema
- [ ] 11.4 Add `InviteResponse`, `InvitePreviewResponse`, `MemberWithProfileResponse` schemas

## 12. Frontend API Layer (Invites + Members)

- [x] 12.1 Add invite types to `frontend/src/types/index.ts`: `WorkspaceInvite`, `InvitePreview`, `MemberWithProfile`
- [x] 12.2 Update `Workspace` type to include `member_count` field
- [x] 12.3 Create `frontend/src/api/invites.ts` with hooks: `usePreviewInvite`, `useAcceptInvite`, `useCreateInvite`
- [x] 12.4 Create `frontend/src/api/members.ts` with hooks: `useListMembers`, `useUpdateMemberRole`, `useRemoveMember`

## 13. Frontend Invite Acceptance Page

- [x] 13.1 Create `frontend/src/pages/InvitePage.tsx` — displays invite preview, handles join/login flow
- [x] 13.2 Add `/invite/:code` route to `App.tsx` (public route, outside AuthGuard)
- [x] 13.3 Implement localStorage bridge: store `sw_pending_invite` before SSO redirect
- [x] 13.4 Update `AuthCallbackPage.tsx` to check for `sw_pending_invite` after successful auth and auto-accept

## 14. Frontend Members Management UI

- [x] 14.1 Create `MembersSection` component for Settings page — list members with avatars, names, roles
- [x] 14.2 Add role change dropdown for owner (calls `useUpdateMemberRole`)
- [x] 14.3 Add remove member button with confirmation dialog for owner (calls `useRemoveMember`)
- [x] 14.4 Create `InviteDialog` component — role selector + generate link + copy URL
- [x] 14.5 Integrate `MembersSection` and invite button into `WorkspaceSettingsPage.tsx`

## 15. Frontend Workspace Switcher Enhancement

- [x] 15.1 Update Header workspace dropdown to show member count per workspace
- [x] 15.2 Add "+ Create workspace" option at bottom of dropdown with separator
- [x] 15.3 Create `CreateWorkspaceDialog` component (name + description fields)
- [x] 15.4 Wire up dialog to `useCreateWorkspace` and `setWorkspaceId` on success
