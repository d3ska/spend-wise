## ADDED Requirements

### Requirement: Split mode selector
The create transaction dialog SHALL provide a split mode selector with three options: "All on me", "Split equally", "Custom". The default mode SHALL be "All on me".

#### Scenario: Default mode on dialog open
- **WHEN** the create transaction dialog opens
- **THEN** the split mode SHALL be set to "All on me"

#### Scenario: User switches mode
- **WHEN** the user selects a different split mode
- **THEN** the splits SHALL be recalculated based on the new mode and current total amount

### Requirement: All on me mode
When split mode is "All on me", the full transaction amount SHALL be assigned to the current user as a single entry. No split detail UI SHALL be displayed.

#### Scenario: Submit with all-on-me
- **WHEN** the user submits with split mode "All on me"
- **THEN** a single entry SHALL be created with `participant_id` set to the current user and the full amount

#### Scenario: No extra UI shown
- **WHEN** split mode is "All on me"
- **THEN** no participant list or split amount fields SHALL be visible

### Requirement: Split equally mode
When split mode is "Split equally", the total amount SHALL be divided equally among all workspace members. The split breakdown SHALL be displayed as a read-only list.

#### Scenario: Equal division
- **WHEN** the user selects "Split equally" and the total amount is set
- **THEN** each workspace member SHALL be shown with their equal share of the amount

#### Scenario: Rounding remainder
- **WHEN** the total amount does not divide evenly (e.g. 100 / 3)
- **THEN** the remainder cents SHALL be assigned to the first participant

#### Scenario: Submit with equal splits
- **WHEN** the user submits with "Split equally" mode
- **THEN** one entry per workspace member SHALL be created, each with `participant_id` and their calculated amount

### Requirement: Custom split mode
When split mode is "Custom", the user SHALL be able to set a specific amount for each workspace member via editable input fields.

#### Scenario: Editable amounts
- **WHEN** the user selects "Custom" mode
- **THEN** each workspace member SHALL be shown with an editable amount input field

#### Scenario: Initial custom values
- **WHEN** the user switches to "Custom" mode
- **THEN** the amounts SHALL be pre-filled with equal splits as a starting point

#### Scenario: Validation mismatch
- **WHEN** custom split amounts do not sum to the total transaction amount (within 0.01 tolerance)
- **THEN** an inline error message SHALL be displayed and the submit button SHALL be disabled

#### Scenario: Submit with custom splits
- **WHEN** the user submits with valid custom splits
- **THEN** one entry per participant with a non-zero amount SHALL be created

### Requirement: Workspace members as participants
The dialog SHALL fetch workspace members to use as split participants.

#### Scenario: Members loaded
- **WHEN** the dialog opens
- **THEN** workspace members SHALL be fetched via the workspace detail endpoint

#### Scenario: Current user label
- **WHEN** displaying member names
- **THEN** the current user SHALL be labeled with their display name followed by "(you)"

### Requirement: Splits terminology
The UI SHALL use the term "Splits" for the cost distribution section.

#### Scenario: Label display
- **WHEN** the split mode selector is rendered
- **THEN** the section label SHALL read "Splits"
