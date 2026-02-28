ALTER TABLE categorization_rules ADD COLUMN bank_account_id BIGINT REFERENCES bank_accounts(id) ON DELETE SET NULL;
