CREATE TABLE bank_connections (
    id               BIGSERIAL   PRIMARY KEY,
    user_id          BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    institution_id   TEXT        NOT NULL,
    institution_name TEXT        NOT NULL,
    session_id       TEXT        NOT NULL UNIQUE,
    status           TEXT        NOT NULL DEFAULT 'active',
    auth_expires_at  TIMESTAMPTZ NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bank_connections_user ON bank_connections (user_id);
CREATE INDEX idx_bank_connections_status ON bank_connections (status);
