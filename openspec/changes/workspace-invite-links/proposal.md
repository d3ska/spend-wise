## Why

SpendWise is a collaborative household ledger, but there's no way to actually invite someone to a workspace. The `AddMember` endpoint requires knowing the target user's numeric ID, and the target must already have an account. For a household app where you want to invite your spouse, this is unusable. We need a simple invite link flow where the owner generates a link, shares it, and the recipient joins by clicking it — whether or not they already have an account.

Additionally, the frontend currently has no UI for managing workspace members, no member count visibility, and no way to create additional workspaces from the main UI.

## What Changes

- **Invite link system**: Owners generate single-use invite links with a role (viewer/editor) and configurable expiration (default 7 days). Recipients click the link to join, going through SSO if not yet logged in.
- **New `workspace_invites` table**: Stores invite codes, target role, expiration, and usage tracking.
- **New API endpoints**: Create invite, preview invite (public), accept invite.
- **Frontend invite page**: `/invite/{code}` route showing workspace preview + join/login flow. Uses localStorage to persist invite code across SSO redirect.
- **Members management UI**: Settings page gets a Members section — list members with profiles, change roles, remove members. Owner-only controls.
- **Workspace switcher enhancement**: Header dropdown shows member count per workspace and a "+ Create workspace" option.
- **Remove `WorkspaceType`**: **BREAKING** — Already removed from backend in prior commit. Frontend types and `useCreateWorkspace` still reference `type` field — clean those up.
- **Enriched ListMembers response**: Existing endpoint returns only IDs; needs to join with users table to return display names, emails, avatars.

## Capabilities

### New Capabilities
- `workspace-invites`: Invite link generation, preview, acceptance, and expiration. Covers the `workspace_invites` table, invite CRUD endpoints, and the frontend invite acceptance flow.
- `member-management-ui`: Frontend UI for listing workspace members with profiles, changing roles, removing members, and generating invite links. Includes workspace switcher enhancements (member count, create workspace).

### Modified Capabilities
- `workspace-management`: Remove `WorkspaceType` from spec (already removed from code). Add enriched member listing that returns user profiles alongside membership data. Add role update capability.

## Impact

- **Database**: New `workspace_invites` table (new migration). Existing `workspace_members` table unchanged.
- **Backend**: New invite handler + service + store. Modified workspace store (enriched member query). Modified workspace service (update member role).
- **Frontend**: New `/invite/{code}` route + page. Modified Settings page (members section). Modified Header (workspace switcher). Modified types (remove `type` from Workspace). Modified `useCreateWorkspace` (remove `type` param).
- **API**: 3 new endpoints (`POST /invites`, `GET /invites/{code}`, `POST /invites/{code}/accept`). 1 modified endpoint (enriched `GET /workspaces/{id}/members`). 1 new endpoint (`PUT /workspaces/{id}/members/{userId}` for role update).
- **OpenAPI spec**: Updated with new invite endpoints and modified member responses.
