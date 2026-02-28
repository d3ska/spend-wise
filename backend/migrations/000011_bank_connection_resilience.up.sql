-- Add iban column to transactions for permanent source account tracking.
ALTER TABLE transactions ADD COLUMN iban TEXT;

-- Backfill iban from existing linked bank accounts.
UPDATE transactions t
SET iban = ba.iban
FROM bank_accounts ba
WHERE t.bank_account_id = ba.id AND ba.iban IS NOT NULL;

-- Index for IBAN+currency lookups during reconnect account reuse.
CREATE INDEX idx_bank_accounts_iban_currency ON bank_accounts (iban, currency) WHERE iban IS NOT NULL;
