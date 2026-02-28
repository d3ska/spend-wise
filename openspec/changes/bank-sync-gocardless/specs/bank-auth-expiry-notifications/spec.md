## ADDED Requirements

### Requirement: Auth expiry status in bank connection response
The bank connection API responses SHALL include computed expiry status fields to enable frontend notification logic.

#### Scenario: Active connection response includes days remaining
- **WHEN** a bank connection with `auth_expires_at` in the future is returned via the API
- **THEN** the response SHALL include `days_remaining` (integer) and `expiry_status` (`"ok"` | `"warning"` | `"expired"`)

#### Scenario: Expiry status thresholds
- **WHEN** `days_remaining > 14`
- **THEN** `expiry_status` SHALL be `"ok"`

#### Scenario: Warning status threshold
- **WHEN** `days_remaining <= 14` and `days_remaining > 0`
- **THEN** `expiry_status` SHALL be `"warning"`

#### Scenario: Expired status
- **WHEN** `days_remaining <= 0`
- **THEN** `expiry_status` SHALL be `"expired"`

### Requirement: Frontend settings page bank connection status
The workspace settings page SHALL display each connected bank account with its auth expiry status.

#### Scenario: Display connected bank with expiry date
- **WHEN** the workspace settings page loads and bank accounts are linked
- **THEN** each bank connection SHALL display: institution name, account name/IBAN, last synced time, and auth expiry date

#### Scenario: Green status for healthy connections
- **WHEN** a bank connection has `expiry_status = "ok"`
- **THEN** it SHALL display with a green indicator and text like "Connected - expires {date}"

#### Scenario: Yellow warning for expiring connections
- **WHEN** a bank connection has `expiry_status = "warning"`
- **THEN** it SHALL display with a yellow indicator and text like "Expiring soon - {days_remaining} days left"
- **AND** a "Reconnect" button SHALL be prominently displayed

#### Scenario: Red status for expired connections
- **WHEN** a bank connection has `expiry_status = "expired"`
- **THEN** it SHALL display with a red indicator and text like "Expired - reconnect to resume syncing"
- **AND** a "Reconnect" button SHALL be prominently displayed

### Requirement: Sidebar warning banner for expiring connections
The application sidebar SHALL display a warning banner when any bank connection linked to the current workspace is expiring soon.

#### Scenario: Warning banner appears at 14 days
- **WHEN** any bank connection linked to the current workspace has `expiry_status = "warning"`
- **THEN** a yellow banner SHALL appear in the sidebar with text like "Bank connection expiring soon" and a link to workspace settings

#### Scenario: No banner when all connections healthy
- **WHEN** all bank connections linked to the current workspace have `expiry_status = "ok"`
- **THEN** no expiry warning banner SHALL be displayed

### Requirement: Expired connection modal on dashboard
The dashboard page SHALL display a modal when any bank connection linked to the current workspace has expired.

#### Scenario: Modal appears for expired connections
- **WHEN** the dashboard loads and any linked bank connection has `expiry_status = "expired"`
- **THEN** a modal SHALL appear listing the expired connections with institution names
- **AND** the modal SHALL include a "Reconnect" button that navigates to workspace settings

#### Scenario: Modal can be dismissed
- **WHEN** the user dismisses the expired connection modal
- **THEN** it SHALL not appear again for that session (until page reload)
