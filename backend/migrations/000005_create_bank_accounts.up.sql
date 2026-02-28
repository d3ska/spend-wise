CREATE TABLE bank_accounts (
    id                 BIGSERIAL   PRIMARY KEY,
    bank_connection_id BIGINT      NOT NULL REFERENCES bank_connections(id) ON DELETE CASCADE,
    external_id        TEXT        NOT NULL UNIQUE,
    iban               TEXT,
    currency           TEXT        NOT NULL DEFAULT '',
    name               TEXT        NOT NULL DEFAULT '',
    last_synced_at     TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bank_accounts_connection ON bank_accounts (bank_connection_id);
