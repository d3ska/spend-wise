ALTER TABLE categorization_rules ADD COLUMN amount_min NUMERIC(19,4);
ALTER TABLE categorization_rules ADD COLUMN amount_max NUMERIC(19,4);
ALTER TABLE categorization_rules ADD CONSTRAINT chk_amount_range
    CHECK (amount_min IS NULL OR amount_max IS NULL OR amount_min <= amount_max);
