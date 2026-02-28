## MODIFIED Requirements

### Requirement: Fingerprint-based deduplication for imports
The system SHALL prevent duplicate transaction imports using a fingerprint hash with a partial unique index. This applies to both CSV imports (source `"import"`) and bank-synced transactions (source `"bank"`).

#### Scenario: Check fingerprint before import
- **WHEN** a transaction with a fingerprint is imported
- **AND** a transaction with the same fingerprint already exists in the workspace
- **THEN** it SHALL be skipped (not inserted)
- **AND** the skip SHALL be counted in the import result

#### Scenario: New fingerprint is accepted
- **WHEN** a transaction with a fingerprint that does not exist in the workspace is imported
- **THEN** it SHALL be created with `source = "import"`

#### Scenario: Manual transactions have no fingerprint
- **WHEN** a transaction is created manually (not via import)
- **THEN** its fingerprint SHALL be NULL
- **AND** it SHALL not be subject to fingerprint deduplication

#### Scenario: Bank-synced transaction uses internalTransactionId as fingerprint
- **WHEN** a transaction is synced from GoCardless
- **THEN** its fingerprint SHALL be set to the GoCardless `internalTransactionId`
- **AND** its source SHALL be `"bank"`
- **AND** if a transaction with the same fingerprint already exists in the workspace, it SHALL be skipped

#### Scenario: Bank sync deduplication across multiple syncs
- **WHEN** the same GoCardless transaction is encountered in consecutive sync runs (due to date overlap buffer)
- **THEN** it SHALL be skipped on subsequent runs via the existing fingerprint unique index
