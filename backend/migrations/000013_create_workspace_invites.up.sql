CREATE TABLE workspace_invites (
    id           BIGSERIAL      PRIMARY KEY,
    workspace_id BIGINT         NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    code         TEXT           NOT NULL UNIQUE,
    role         member_role    NOT NULL,
    created_by   BIGINT         NOT NULL REFERENCES users(id),
    expires_at   TIMESTAMPTZ    NOT NULL,
    used_by      BIGINT         REFERENCES users(id),
    used_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
