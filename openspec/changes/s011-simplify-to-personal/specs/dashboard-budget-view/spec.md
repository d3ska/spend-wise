## ADDED Requirements

### Requirement: Summary endpoint returns budget data
The GET `/api/v1/workspaces/{id}/summary` endpoint SHALL include budget information in the response when budgets are defined for the queried date range:
- `overall_budget`: Money object or `null` if no overall budget is set
- Each entry in `by_category` SHALL include a `budget` field: Money object or `null`

Budget amounts SHALL be summed across all months in the queried date range (e.g., if querying Jan-Mar and each month has a 1000 PLN overall budget, `overall_budget` returns 3000 PLN).

#### Scenario: Summary with overall and category budgets
- **WHEN** a summary is requested for a date range where an overall budget of 5000 PLN and a Groceries budget of 600 PLN exist for one month
- **THEN** the response SHALL include `"overall_budget": { "amount": "5000.00", "currency": "PLN" }`
- **AND** the Groceries entry in `by_category` SHALL include `"budget": { "amount": "600.00", "currency": "PLN" }`
- **AND** categories without budgets SHALL have `"budget": null`

#### Scenario: Summary without any budgets
- **WHEN** a summary is requested and no funding/budget records exist for the date range
- **THEN** `overall_budget` SHALL be `null`
- **AND** all `by_category` entries SHALL have `"budget": null`

#### Scenario: Multi-month budget aggregation
- **WHEN** a summary is requested for Jan-Mar 2026 and the user has set a 2000 PLN overall budget for each of the three months
- **THEN** `overall_budget` SHALL be `{ "amount": "6000.00", "currency": "PLN" }` (sum of 3 months)

### Requirement: Dashboard overall budget widget
When `overall_budget` is present in the summary, the dashboard SHALL replace the plain "Total Spent" card with a budget progress widget that shows:
- The amount spent and the budget amount (e.g., "1,234 / 5,000 PLN")
- A progress bar representing spent / budget ratio
- Color coding: green when under 75%, yellow between 75-100%, red when over 100%

#### Scenario: Under budget
- **WHEN** the dashboard loads with total spent = 2,500 PLN and overall budget = 5,000 PLN
- **THEN** the progress bar SHALL show 50% filled with green color
- **AND** the display SHALL show "2,500.00 / 5,000.00 PLN"

#### Scenario: Near budget limit
- **WHEN** total spent = 4,200 PLN and overall budget = 5,000 PLN
- **THEN** the progress bar SHALL show 84% filled with yellow color

#### Scenario: Over budget
- **WHEN** total spent = 5,800 PLN and overall budget = 5,000 PLN
- **THEN** the progress bar SHALL show 100% filled (capped visually) with red color
- **AND** the display SHALL indicate overspend (e.g., "5,800.00 / 5,000.00 PLN")

#### Scenario: No budget defined
- **WHEN** the dashboard loads with `overall_budget = null`
- **THEN** the "Total Spent" card SHALL display as it does today (plain amount, no progress bar)

### Requirement: Dashboard category budget bars
When any category in `by_category` has a non-null `budget`, the dashboard SHALL display a category budget section showing each budgeted category with:
- Category name and icon
- Spent vs budget amounts
- A horizontal progress bar with the same color coding as the overall widget (green < 75%, yellow 75-100%, red > 100%)

Categories without budgets SHALL not appear in this section. If no categories have budgets, this section SHALL not render.

#### Scenario: Mixed categories with and without budgets
- **WHEN** the summary has 4 categories but only 2 have budgets defined
- **THEN** the category budget section SHALL show only the 2 budgeted categories with progress bars
- **AND** the other 2 categories SHALL still appear in the pie chart but not in the budget section

#### Scenario: Category over budget
- **WHEN** Groceries has spent 700 PLN against a 600 PLN budget
- **THEN** the Groceries bar SHALL display red with "700.00 / 600.00 PLN"

#### Scenario: No categories have budgets
- **WHEN** `overall_budget` is set but no per-category budgets exist
- **THEN** the category budget section SHALL not render
- **AND** the overall budget widget SHALL still display
