## ADDED Requirements

### Requirement: Fingerprint-based deduplication for imports
The system SHALL prevent duplicate transaction imports using a fingerprint hash with a partial unique index.

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

### Requirement: Import endpoint accepts batch of transactions
The import endpoint SHALL accept a JSON array of transaction inputs and return counts of imported and skipped transactions.

#### Scenario: Successful import with some duplicates
- **WHEN** `POST /api/v1/workspaces/{id}/transactions/import` is called with 5 transactions, 2 of which have existing fingerprints
- **THEN** the response SHALL contain `imported: 3`, `skipped: 2`
- **AND** the 3 new transactions SHALL be returned in the response

#### Scenario: Import requires editor role
- **WHEN** a viewer attempts to import transactions
- **THEN** the response SHALL be `403 Forbidden`
