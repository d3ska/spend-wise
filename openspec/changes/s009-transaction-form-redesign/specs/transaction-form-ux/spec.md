## ADDED Requirements

### Requirement: Top-level category field
The create transaction form SHALL display category as a single top-level select field, not repeated per split/entry.

#### Scenario: User selects category once
- **WHEN** user opens the create transaction dialog
- **THEN** a single category dropdown is displayed at the top level of the form, outside of the splits section

#### Scenario: Category applies to all entries
- **WHEN** user submits the form with a selected category
- **THEN** all generated entries SHALL use the same `category_id`

### Requirement: Auto-suggest category from rules
The form SHALL auto-suggest a category based on the typed description by matching against enabled workspace rules client-side.

#### Scenario: Description matches a rule pattern
- **WHEN** user types a description that contains the `match_pattern` of an enabled rule (case-insensitive)
- **THEN** the category dropdown SHALL auto-select the `target_category_id` of the highest-priority matching rule
- **AND** the user MAY override the suggestion by selecting a different category

#### Scenario: No rule matches
- **WHEN** user types a description that does not match any enabled rule
- **THEN** the category dropdown SHALL remain unchanged (no auto-selection)

#### Scenario: Rules not loaded yet
- **WHEN** rules data is still loading
- **THEN** auto-suggest SHALL be disabled and the category dropdown works as a normal select

### Requirement: Splits terminology
The UI SHALL use the term "Splits" instead of "Entries" for the participant cost distribution section.

#### Scenario: Label display
- **WHEN** the create transaction dialog is rendered
- **THEN** all labels, buttons, and headings SHALL use "Splits" (e.g., "Add split") not "Entries"

### Requirement: Participant-based splits
Each split SHALL represent a workspace member and their share of the transaction amount.

#### Scenario: Default split
- **WHEN** user opens the create transaction dialog
- **THEN** the split mode SHALL default to "All on me" with 100% of the amount assigned to the current user

#### Scenario: Split equally mode
- **WHEN** user selects "Split equally" mode
- **THEN** the total amount SHALL be divided equally among all workspace members
- **AND** any rounding remainder (cents) SHALL be assigned to the first participant

#### Scenario: Custom split mode
- **WHEN** user selects "Custom" mode
- **THEN** the user SHALL be able to set a specific amount for each workspace member
- **AND** the form SHALL validate that split amounts sum to the total amount

### Requirement: Split mode selector
The form SHALL provide a split mode selector with three options: "All on me", "Split equally", "Custom".

#### Scenario: Mode switching
- **WHEN** user changes the split mode
- **THEN** the splits SHALL be recalculated based on the new mode and current total amount

#### Scenario: All on me hides split details
- **WHEN** split mode is "All on me"
- **THEN** no individual split rows SHALL be displayed (implicit single split)

#### Scenario: Equal and custom show member list
- **WHEN** split mode is "Split equally" or "Custom"
- **THEN** a list of workspace members with their split amounts SHALL be displayed

### Requirement: Workspace members for splits
The form SHALL fetch workspace members to populate split participants.

#### Scenario: Members loaded
- **WHEN** the create transaction dialog opens
- **THEN** workspace members SHALL be fetched via the workspace detail endpoint
- **AND** each member SHALL be available as a split participant with their display name
