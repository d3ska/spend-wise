## ADDED Requirements

### Requirement: Summary stat cards
The dashboard SHALL display summary cards showing key financial metrics from the `GET /summary` endpoint.

#### Scenario: Summary cards rendered
- **WHEN** the dashboard loads with summary data
- **THEN** cards SHALL display "Total Spent" and the date range for the current period

#### Scenario: Money formatting
- **WHEN** a money value is displayed
- **THEN** it SHALL be formatted with currency symbol, thousands separator, and two decimal places (e.g., "$1,234.56")

### Requirement: Category breakdown chart
The dashboard SHALL display a pie or donut chart showing spending by category using Recharts.

#### Scenario: Chart renders with data
- **WHEN** the summary contains `by_category` data with multiple categories
- **THEN** a chart SHALL render with each category as a segment, labeled with name and percentage

#### Scenario: Empty state
- **WHEN** the summary contains no spending data
- **THEN** the chart area SHALL show a friendly empty state message (e.g., "No spending data yet")

### Requirement: Recent transactions list
The dashboard SHALL display a brief list of recent transactions.

#### Scenario: Recent transactions shown
- **WHEN** the dashboard loads
- **THEN** it SHALL display the 5 most recent transactions with date, description, category, and amount

#### Scenario: Transaction link
- **WHEN** the user clicks on a transaction in the recent list
- **THEN** they SHALL be navigated to the transactions page

### Requirement: Date range selector
The dashboard SHALL allow the user to select a date range for the summary data.

#### Scenario: Default date range
- **WHEN** the dashboard loads without a selected range
- **THEN** the default range SHALL be the current calendar month

#### Scenario: Custom date range
- **WHEN** the user selects a custom date range
- **THEN** the summary data and charts SHALL update to reflect the selected range
