CREATE TYPE member_role AS ENUM ('owner', 'editor', 'viewer');

CREATE TABLE workspaces (
    id          BIGSERIAL      PRIMARY KEY,
    name        TEXT           NOT NULL,
    description TEXT           NOT NULL,
    owner_id    BIGINT         NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE workspace_members (
    workspace_id BIGINT      NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id      BIGINT      NOT NULL REFERENCES users(id),
    role         member_role NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (workspace_id, user_id)
);
