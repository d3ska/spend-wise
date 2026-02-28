## Context

SpendWise workspaces support multiple members via `workspace_members`, but there's no way to invite someone. The `AddMember` endpoint takes a numeric `user_id`, requiring the target to already have an account and the inviter to know their ID. The frontend has no member management UI — no member list, no role changes, no invite flow.

The `WorkspaceType` (private/shared) was just removed from the backend — it was unused indirection since the membership system already determines sharing. The frontend still has stale `type` references to clean up.

The auth flow uses SSO (Google/GitHub) via OAuth2 with httpOnly cookie JWT sessions. Users are created via `GetOrCreateByEmail` on first login.

## Goals / Non-Goals

**Goals:**
- Single-use invite links with configurable expiration and role assignment
- Seamless join flow for users who aren't logged in yet (SSO bridge)
- Member management UI: list, role changes, removal
- Workspace switcher showing member counts + ability to create new workspaces
- Clean up stale `WorkspaceType` references in frontend

**Non-Goals:**
- Email-based invitations (sending emails) — link sharing is manual
- Invite notifications or in-app notification system
- Bulk invites or team management
- Workspace discovery or public workspaces
- Invite revocation UI (expired/used invites are effectively dead)

## Decisions

### 1. Single-use invite codes stored in database

**Decision**: `workspace_invites` table with crypto-random URL-safe codes (32 bytes, base64url). Single-use (`max_uses=1`), configurable expiry defaulting to 7 days.

**Alternatives considered**:
- JWT-based invite tokens (no DB storage): Simpler but can't revoke, can't track usage, can't enforce single-use server-side.
- Email-based invites with pending state: Requires email infrastructure; overkill for household use.

**Rationale**: DB-backed codes are simple, auditable, and give us server-side enforcement of single-use + expiration.

### 2. Frontend localStorage for SSO bridge

**Decision**: When an unauthenticated user visits `/invite/{code}`, the frontend stores the code in `localStorage` before redirecting to login. After SSO callback, the frontend checks for a pending invite and calls the accept endpoint.

**Alternatives considered**:
- Encode invite code in OAuth state parameter: Couples auth and invite concerns; state param has a CSRF purpose.
- Server-side session cookie: Adds server-side state management; the backend is currently stateless (JWT only).

**Rationale**: Auth and invite acceptance are two independent operations. The frontend orchestrates the sequence. This is the standard pattern (Slack, Notion, Discord) — composable, resilient to SSO failures, works for already-logged-in users too.

### 3. Enriched member listing via SQL JOIN

**Decision**: Modify the `ListMembers` sqlc query to JOIN `workspace_members` with `users` to return `display_name`, `email`, `avatar_url` alongside membership data. New sqlc query, not modifying the existing one.

**Alternatives considered**:
- Two separate API calls (list members, then fetch each user): N+1 problem, chatty.
- Return user IDs and let frontend resolve: Shifts complexity to client.

**Rationale**: A single query with JOIN is the simplest and most efficient approach.

### 4. Role update via dedicated endpoint

**Decision**: `PUT /workspaces/{id}/members/{userId}` with `{ "role": "editor" }`. Owner-only. Cannot change own role. Cannot set role to owner (ownership transfer is a separate concern).

**Alternatives considered**:
- PATCH with partial updates: Overengineered for a single-field change.

### 5. Invite preview is public, accept requires auth

**Decision**: `GET /invites/{code}` is a public endpoint returning workspace name and inviter name (minimal info). `POST /invites/{code}/accept` requires authentication (JWT). This lets the frontend show a preview page before login.

### 6. New migration file for invites table

**Decision**: New migration `000010_create_workspace_invites` (next available number). The `workspace_invites` table references `workspaces(id)` with CASCADE delete.

## Risks / Trade-offs

- **Invite link guessing**: 32 bytes of crypto random → 256 bits of entropy. Effectively unguessable. No rate limiting needed for preview endpoint.
- **Stale invites accumulating**: Expired invites stay in DB. Low risk for household-scale usage. Could add periodic cleanup later if needed.
- **localStorage cleared**: If user clears browser data between clicking invite and completing SSO, the invite code is lost. They'd need the link again. Acceptable for household use.
- **No invite revocation UI**: Owner can't cancel a sent invite. Single-use + 7-day expiry mitigates this. Could add later.
- **Owner role transfer not supported**: `UpdateMemberRole` explicitly rejects setting role to "owner". This is intentional — ownership transfer is a separate, more dangerous operation.

## Migration Plan

1. Add new migration for `workspace_invites` table
2. Add new sqlc queries and regenerate
3. Implement backend (store → service → handler)
4. Add frontend invite flow + member management UI
5. Clean up stale `WorkspaceType` references in frontend
6. No data migration needed — this is purely additive
