## ADDED Requirements

### Requirement: Categories list with CRUD
The categories page SHALL display all categories for the current workspace and allow creating, editing, and deleting categories.

#### Scenario: List categories
- **WHEN** the categories page loads
- **THEN** it SHALL display all categories with their name and icon

#### Scenario: Create category
- **WHEN** the user fills in a category name and icon and submits
- **THEN** the app SHALL POST to the create category endpoint and refresh the list

#### Scenario: Edit category
- **WHEN** the user clicks edit on a category
- **THEN** an edit form SHALL appear with the current name and icon pre-filled
- **AND** on submit, the app SHALL PUT to the update endpoint and refresh the list

#### Scenario: Delete category
- **WHEN** the user clicks delete on a category
- **THEN** a confirmation dialog SHALL appear
- **AND** on confirm, the app SHALL call the delete endpoint and refresh the list
