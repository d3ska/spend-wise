CREATE TABLE fundings (
    id           BIGSERIAL     PRIMARY KEY,
    workspace_id BIGINT        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id      BIGINT        NOT NULL REFERENCES users(id),
    year_month   TEXT          NOT NULL,
    amount       NUMERIC(19,4) NOT NULL,
    currency     TEXT          NOT NULL DEFAULT 'PLN',
    category_id  BIGINT        REFERENCES categories(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX fundings_workspace_user_month_category_idx
    ON fundings (workspace_id, user_id, year_month, COALESCE(category_id, 0));
