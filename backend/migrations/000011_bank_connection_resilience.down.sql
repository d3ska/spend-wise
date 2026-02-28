DROP INDEX IF EXISTS idx_bank_accounts_iban_currency;
ALTER TABLE transactions DROP COLUMN IF EXISTS iban;
