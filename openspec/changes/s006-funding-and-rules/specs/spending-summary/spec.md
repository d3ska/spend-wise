## MODIFIED Requirements

### Requirement: Funded and balance fields
Each per-participant item SHALL include `funded` and `balance` fields. The `funded` amount SHALL be computed by summing the participant's funding records for the months overlapping the summary date range. The `balance` SHALL equal `funded - spent`.

#### Scenario: Participant with funding
- **WHEN** participant A has spent 120.00 and funded 300.00 in the overlapping months
- **THEN** `funded` SHALL be 300.00 and `balance` SHALL be 180.00

#### Scenario: Participant with no funding
- **WHEN** participant B has spent 80.00 and has no funding records
- **THEN** `funded` SHALL be 0.00 and `balance` SHALL be -80.00

#### Scenario: Multiple months in range
- **WHEN** the date range spans February and March, and a participant funded 200.00 in February and 150.00 in March
- **THEN** `funded` SHALL be 350.00
