## ADDED Requirements

### Requirement: Auto-categorization during import
When transactions are imported, the system SHALL attempt to auto-categorize each transaction by resolving its description against the tiered rule engine. If a matching rule is found and the transaction's entries do not already have a category assigned, the resolved category SHALL be applied.

#### Scenario: Rule matches imported transaction
- **WHEN** a transaction with description "Biedronka grocery store" is imported
- **AND** a workspace rule with pattern "biedronka" targeting category "Groceries" exists and is enabled
- **THEN** the transaction's entries SHALL be assigned to the "Groceries" category

#### Scenario: No rule matches
- **WHEN** a transaction description matches no enabled rule
- **THEN** the import SHALL proceed without auto-categorization (entries keep their provided category)

#### Scenario: Disabled rule does not match
- **WHEN** a matching rule exists but is disabled
- **THEN** it SHALL NOT be applied during import
