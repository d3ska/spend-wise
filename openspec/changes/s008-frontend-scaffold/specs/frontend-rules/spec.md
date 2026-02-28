## ADDED Requirements

### Requirement: Rules list with CRUD and toggle
The rules page SHALL display all rules for the current workspace and allow creating, editing, deleting, and toggling rules.

#### Scenario: List rules
- **WHEN** the rules page loads
- **THEN** it SHALL display all rules with their name, conditions, actions, and enabled/disabled status

#### Scenario: Create rule
- **WHEN** the user fills in the rule form and submits
- **THEN** the app SHALL POST to the create rule endpoint and refresh the list

#### Scenario: Edit rule
- **WHEN** the user clicks edit on a rule
- **THEN** an edit form SHALL appear pre-filled with the rule's current values
- **AND** on submit, the app SHALL PUT to the update endpoint and refresh the list

#### Scenario: Delete rule
- **WHEN** the user clicks delete on a rule
- **THEN** a confirmation dialog SHALL appear
- **AND** on confirm, the app SHALL call the delete endpoint and refresh the list

#### Scenario: Toggle rule
- **WHEN** the user clicks the toggle switch on a rule
- **THEN** the app SHALL PATCH to the toggle endpoint
- **AND** the rule's enabled/disabled state SHALL update in the UI
