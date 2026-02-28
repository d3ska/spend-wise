## ADDED Requirements

### Requirement: Fundings list and record form
The fundings page SHALL display all fundings for the current workspace and allow recording new fundings.

#### Scenario: List fundings
- **WHEN** the fundings page loads
- **THEN** it SHALL display all fundings with date, member, amount, and description

#### Scenario: Record funding
- **WHEN** the user fills in amount, member, and optional description and submits
- **THEN** the app SHALL PUT to the record funding endpoint and refresh the list

#### Scenario: Empty state
- **WHEN** there are no fundings
- **THEN** the page SHALL show "No fundings recorded yet" with a prompt to record one
