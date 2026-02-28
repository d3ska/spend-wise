## ADDED Requirements

### Requirement: App shell with sidebar and header
The app SHALL render an app shell with a fixed sidebar on the left and a header bar at the top for all authenticated pages.

#### Scenario: Desktop layout
- **WHEN** the viewport width is >= 768px
- **THEN** the sidebar SHALL be visible with navigation links: Dashboard, Transactions, Categories, Fundings, Rules, Settings

#### Scenario: Active route highlighted
- **WHEN** the user is on the Transactions page
- **THEN** the "Transactions" link in the sidebar SHALL be visually highlighted

### Requirement: Responsive sidebar collapse
The sidebar SHALL collapse to a hamburger menu on mobile viewports.

#### Scenario: Mobile layout
- **WHEN** the viewport width is < 768px
- **THEN** the sidebar SHALL be hidden
- **AND** a hamburger menu button SHALL appear in the header

#### Scenario: Mobile menu opens
- **WHEN** the user taps the hamburger button on mobile
- **THEN** the sidebar SHALL slide in as an overlay with the same navigation links

### Requirement: Header with workspace selector and user menu
The header SHALL display the app name, a workspace selector dropdown, and a user avatar with a dropdown menu.

#### Scenario: Workspace selector
- **WHEN** the user clicks the workspace selector
- **THEN** a dropdown SHALL show all workspaces the user is a member of
- **AND** selecting a workspace SHALL switch the active workspace context

#### Scenario: User menu
- **WHEN** the user clicks their avatar in the header
- **THEN** a dropdown SHALL show their name, email, and a "Log out" option

### Requirement: First-time workspace onboarding
When a user has no workspaces, the app SHALL prompt them to create one before accessing other features.

#### Scenario: No workspaces exist
- **WHEN** an authenticated user has zero workspaces
- **THEN** the app SHALL show a "Create your first workspace" prompt instead of the dashboard

#### Scenario: Single workspace auto-selected
- **WHEN** a user has exactly one workspace
- **THEN** it SHALL be auto-selected as the active workspace
