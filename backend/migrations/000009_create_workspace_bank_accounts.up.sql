CREATE TABLE workspace_bank_accounts (
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    bank_account_id BIGINT NOT NULL REFERENCES bank_accounts(id) ON DELETE CASCADE,
    UNIQUE (workspace_id, bank_account_id)
);

CREATE INDEX idx_workspace_bank_accounts_bank ON workspace_bank_accounts (bank_account_id);
