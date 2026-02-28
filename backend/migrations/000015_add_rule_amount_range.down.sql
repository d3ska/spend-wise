ALTER TABLE categorization_rules DROP CONSTRAINT IF EXISTS chk_amount_range;
ALTER TABLE categorization_rules DROP COLUMN IF EXISTS amount_max;
ALTER TABLE categorization_rules DROP COLUMN IF EXISTS amount_min;
