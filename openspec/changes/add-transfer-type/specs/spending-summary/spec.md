## ADDED Requirements

### Requirement: Transfers excluded from spending summaries
The spending summary queries SHALL continue to filter on `transaction_type = 'expense'`, which inherently excludes `transfer` transactions. No query changes are required — this spec documents the expected behavior.

#### Scenario: Transfer not counted in total spent
- **WHEN** a workspace has expenses totaling 500 and transfers totaling 200
- **THEN** `total_spent` SHALL be 500 (transfers excluded)

#### Scenario: Transfer not counted in category breakdown
- **WHEN** a transfer transaction exists with category "Uncategorized"
- **THEN** it SHALL NOT appear in the `by_category` breakdown

#### Scenario: Transfer not counted in participant breakdown
- **WHEN** a transfer transaction has a participant assigned
- **THEN** it SHALL NOT appear in the `by_participant` breakdown
