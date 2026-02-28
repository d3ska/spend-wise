CREATE TYPE rule_scope AS ENUM ('system', 'user', 'workspace');

CREATE TABLE categorization_rules (
    id                 BIGSERIAL   PRIMARY KEY,
    scope              rule_scope  NOT NULL,
    owner_id           BIGINT      REFERENCES users(id),
    workspace_id       BIGINT      REFERENCES workspaces(id) ON DELETE CASCADE,
    match_pattern      TEXT        NOT NULL,
    target_category_id BIGINT      NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    priority           INT         NOT NULL DEFAULT 0,
    enabled            BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_scope_fk CHECK (
        (scope = 'workspace' AND workspace_id IS NOT NULL) OR
        (scope = 'user' AND owner_id IS NOT NULL) OR
        (scope = 'system')
    )
);

CREATE INDEX idx_rules_workspace ON categorization_rules (workspace_id) WHERE workspace_id IS NOT NULL;
CREATE INDEX idx_rules_owner ON categorization_rules (owner_id) WHERE owner_id IS NOT NULL;
