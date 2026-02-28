CREATE TYPE transaction_source AS ENUM ('manual', 'import', 'bank');
CREATE TYPE transaction_type AS ENUM ('expense', 'income');

CREATE TABLE transactions (
    id               BIGSERIAL          PRIMARY KEY,
    workspace_id     BIGINT             NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by       BIGINT             NOT NULL REFERENCES users(id),
    total_amount     NUMERIC(19,4)      NOT NULL,
    currency         TEXT               NOT NULL DEFAULT 'PLN',
    description      TEXT               NOT NULL,
    date             DATE               NOT NULL,
    fingerprint      TEXT,
    source           transaction_source NOT NULL,
    transaction_type transaction_type   NOT NULL DEFAULT 'expense',
    notes            TEXT               NOT NULL DEFAULT '',
    bank_account_id  BIGINT             REFERENCES bank_accounts(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_workspace_date ON transactions (workspace_id, date);
CREATE UNIQUE INDEX idx_transactions_fingerprint ON transactions (workspace_id, fingerprint) WHERE fingerprint IS NOT NULL;

CREATE TABLE entries (
    id             BIGSERIAL     PRIMARY KEY,
    transaction_id BIGINT        NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    category_id    BIGINT        NOT NULL REFERENCES categories(id),
    participant_id BIGINT        REFERENCES users(id),
    amount         NUMERIC(19,4) NOT NULL,
    currency       TEXT          NOT NULL DEFAULT 'PLN',
    note           TEXT          NOT NULL DEFAULT ''
);
