## ADDED Requirements

### Requirement: Spending summary read model
The system SHALL provide a `SpendingSummary` read model containing the workspace ID, date range, total spent amount, a list of per-category spending breakdowns, and a list of per-participant spending breakdowns. The summary SHALL be computed on-the-fly from aggregate SQL queries and SHALL NOT be persisted.

#### Scenario: Summary contains all sections
- **WHEN** a summary is requested for a workspace with transactions in the date range
- **THEN** the response SHALL include `total_spent`, `by_category` (array), and `by_participant` (array)

#### Scenario: Empty workspace returns zero totals
- **WHEN** a summary is requested for a workspace with no transactions in the date range
- **THEN** `total_spent` SHALL be zero, `by_category` SHALL be an empty array, and `by_participant` SHALL be an empty array

### Requirement: Total spent aggregation
The system SHALL compute the total spent amount by summing all entry amounts from transactions within the specified workspace and date range. The date range filter SHALL use `date >= from AND date < to` (inclusive start, exclusive end).

#### Scenario: Sum across multiple transactions
- **WHEN** three transactions exist with entry totals of 100.00, 50.00, and 25.50 within the date range
- **THEN** `total_spent` SHALL be 175.50

#### Scenario: Transactions outside range excluded
- **WHEN** transactions exist both inside and outside the date range
- **THEN** only transactions with `date >= from AND date < to` SHALL be included in the sum

### Requirement: Per-category spending breakdown
The system SHALL compute spending per category by summing entry amounts grouped by category for transactions within the date range. Each item SHALL include the category ID, category name, category icon, and the spent amount.

#### Scenario: Multiple categories with entries
- **WHEN** entries exist across categories "Food" (80.00), "Transport" (40.00), and "Entertainment" (30.00)
- **THEN** `by_category` SHALL contain three items with the respective spent amounts

#### Scenario: Category with no entries in range
- **WHEN** a category exists in the workspace but has no entries in the date range
- **THEN** that category SHALL NOT appear in the `by_category` array

### Requirement: Per-participant spending breakdown
The system SHALL compute spending per participant by summing entry amounts grouped by participant ID for transactions within the date range. Each item SHALL include the participant's user ID, display name, and spent amount. Entries with no participant (NULL participant_id) SHALL be excluded from the per-participant breakdown.

#### Scenario: Two participants with entries
- **WHEN** participant A has entries totaling 120.00 and participant B has entries totaling 80.00
- **THEN** `by_participant` SHALL contain two items with the respective spent amounts

#### Scenario: Entries without participant excluded
- **WHEN** some entries have NULL participant_id
- **THEN** those entries SHALL be included in `total_spent` but SHALL NOT appear in `by_participant`

### Requirement: Funded and balance fields
Each per-participant item SHALL include `funded` and `balance` fields. Until the funding feature is implemented, `funded` SHALL be zero and `balance` SHALL equal the negation of `spent` (i.e., `0 - spent`).

#### Scenario: Balance without funding
- **WHEN** participant A has spent 120.00 and no funding data exists
- **THEN** `funded` SHALL be 0.00 and `balance` SHALL be -120.00

### Requirement: Date-range agnostic queries
The summary endpoint SHALL accept arbitrary date ranges via `from` and `to` query parameters (format: YYYY-MM-DD). The backend SHALL NOT encode time granularity — the same query shape handles monthly, weekly, daily, and custom ranges.

#### Scenario: Monthly query
- **WHEN** `from=2026-02-01` and `to=2026-03-01`
- **THEN** the summary SHALL cover the entire month of February 2026

#### Scenario: Daily query
- **WHEN** `from=2026-02-12` and `to=2026-02-13`
- **THEN** the summary SHALL cover a single day (February 12)

### Requirement: Permission check
The summary endpoint SHALL require the requesting user to have viewer or higher role in the workspace. Users without membership SHALL receive a permission error.

#### Scenario: Viewer can access summary
- **WHEN** a user with viewer role requests the summary
- **THEN** the summary SHALL be returned successfully

#### Scenario: Non-member denied
- **WHEN** a user who is not a workspace member requests the summary
- **THEN** the system SHALL return a permission error
