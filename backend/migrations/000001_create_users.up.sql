CREATE TYPE auth_provider AS ENUM ('google', 'github');

CREATE TABLE users (
    id           BIGSERIAL PRIMARY KEY,
    email        TEXT        NOT NULL UNIQUE,
    display_name TEXT        NOT NULL,
    avatar_url   TEXT        NOT NULL DEFAULT '',
    provider     auth_provider NOT NULL,
    provider_id  TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
