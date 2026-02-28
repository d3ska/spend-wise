## MODIFIED Requirements

### Requirement: Bank connection entity with GoCardless requisition tracking
The `model` package SHALL define a `BankConnection` entity with typed `BankConnectionID`, user ownership, Enable Banking session reference, institution metadata, status tracking, and auth expiry.

#### Scenario: BankConnection fields
- **WHEN** a `BankConnection` struct is created
- **THEN** it SHALL have fields: `ID BankConnectionID`, `UserID UserID`, `InstitutionID string`, `InstitutionName string`, `SessionID string`, `Status BankConnectionStatus`, `AuthExpiresAt time.Time`, `CreatedAt time.Time`, `UpdatedAt time.Time`

#### Scenario: BankConnectionStatus enum values
- **WHEN** the `BankConnectionStatus` type is defined
- **THEN** it SHALL support values: `BankConnectionActive`, `BankConnectionExpired`, `BankConnectionRevoked`

### Requirement: Bank connection database migration
The database SHALL have a `bank_connections` table storing user-owned Enable Banking session data.

#### Scenario: Bank connections table schema
- **WHEN** the bank connections migration is applied
- **THEN** the `bank_connections` table SHALL have columns: `id` (BIGSERIAL PK), `user_id` (BIGINT NOT NULL FK to users ON DELETE CASCADE), `institution_id` (TEXT NOT NULL), `institution_name` (TEXT NOT NULL), `session_id` (TEXT NOT NULL UNIQUE), `status` (TEXT NOT NULL DEFAULT 'active'), `auth_expires_at` (TIMESTAMPTZ NOT NULL), `created_at` (TIMESTAMPTZ NOT NULL DEFAULT NOW()), `updated_at` (TIMESTAMPTZ NOT NULL DEFAULT NOW())

#### Scenario: Migration from requisition_id to session_id
- **WHEN** the rename migration is applied
- **THEN** the `requisition_id` column SHALL be renamed to `session_id`
- **AND** the unique index SHALL be recreated on the new column name

#### Scenario: Migration rollback
- **WHEN** the rename migration is rolled back
- **THEN** the `session_id` column SHALL be renamed back to `requisition_id`

### Requirement: Bank connection store with CRUD operations
The `store` package SHALL provide persistence for bank connections.

#### Scenario: Create bank connection
- **WHEN** `Create(ctx, model.BankConnection)` is called with valid fields
- **THEN** it SHALL insert a row and return the created `BankConnection` with a generated ID

#### Scenario: Get bank connection by ID
- **WHEN** `GetByID(ctx, id)` is called for an existing connection
- **THEN** it SHALL return the `BankConnection`

#### Scenario: Get bank connection not found
- **WHEN** `GetByID(ctx, id)` is called for a non-existent connection
- **THEN** it SHALL return `model.ErrBankConnectionNotFound`

#### Scenario: List bank connections by user
- **WHEN** `ListByUser(ctx, userID)` is called
- **THEN** it SHALL return all bank connections owned by that user, ordered by `created_at` descending

#### Scenario: List active bank connections for sync
- **WHEN** `ListActive(ctx)` is called
- **THEN** it SHALL return all bank connections with `status = 'active'` and `auth_expires_at > now()`

#### Scenario: Update bank connection status
- **WHEN** `UpdateStatus(ctx, id, status)` is called
- **THEN** it SHALL update the `status` and `updated_at` fields

#### Scenario: Update bank connection session
- **WHEN** `UpdateSession(ctx, id, sessionID, status, authExpiresAt)` is called
- **THEN** it SHALL update the `session_id`, `status`, `auth_expires_at`, and `updated_at` fields

#### Scenario: Delete bank connection
- **WHEN** `Delete(ctx, id, userID)` is called
- **THEN** it SHALL delete the connection only if owned by the given user
- **AND** associated `bank_accounts` and `workspace_bank_accounts` SHALL be removed via CASCADE

### Requirement: Bank connection service with Enable Banking auth flow orchestration
The `service` package SHALL provide bank connection operations using the Enable Banking auth/session lifecycle.

#### Scenario: List available ASPSPs
- **WHEN** `ListInstitutions(ctx, countryCode)` is called
- **THEN** it SHALL delegate to the Enable Banking client and return the ASPSP list

#### Scenario: Initiate bank connection
- **WHEN** `Initiate(ctx, userID, aspspName, aspspCountry, institutionName, redirectURL)` is called
- **THEN** it SHALL query the ASPSP metadata to get `maximum_consent_validity`
- **AND** it SHALL call `StartAuth` with `valid_until` = `min(max_consent_validity, 180 days)` from now
- **AND** it SHALL persist a `BankConnection` with status `active`, `session_id` set to the `authorization_id`, and `auth_expires_at` based on the computed validity
- **AND** it SHALL return the auth URL for user redirect

#### Scenario: Complete bank connection
- **WHEN** `Complete(ctx, userID, connectionID, authorizationCode)` is called
- **THEN** it SHALL call `CreateSession(ctx, code)` to exchange the code for a session
- **AND** it SHALL update the connection's `session_id` to the new session ID
- **AND** it SHALL create `BankAccount` records for each discovered account, populating `external_id` with the account `uid` and `iban` from `account_id`

#### Scenario: Delete bank connection requires ownership
- **WHEN** a user attempts to delete a connection they do not own
- **THEN** it SHALL return `model.ErrBankConnectionNotFound`

#### Scenario: Reconnect expired connection
- **WHEN** `Reconnect(ctx, userID, connectionID, redirectURL)` is called for an expired connection
- **THEN** it SHALL call `StartAuth` for the same ASPSP with a new consent period
- **AND** it SHALL update the connection's `session_id`, `status` to `active`, and `auth_expires_at`

### Requirement: Bank connection HTTP handler
The `handler` package SHALL expose REST endpoints for bank connection management.

#### Scenario: List institutions endpoint
- **WHEN** `GET /api/v1/banks?country={code}` is called
- **THEN** it SHALL return the list of available ASPSPs for that country
- **AND** the `country` query parameter SHALL default to `"PL"` if omitted

#### Scenario: Initiate connection endpoint
- **WHEN** `POST /api/v1/bank-connections` is called with `{"institution_id": "...", "institution_name": "...", "redirect_url": "..."}`
- **THEN** it SHALL return the bank connection ID and the Enable Banking auth URL
- **AND** it SHALL respond with `201 Created`

#### Scenario: Complete connection endpoint
- **WHEN** `POST /api/v1/bank-connections/{id}/complete` is called with the authorization code
- **THEN** it SHALL finalize the connection and return the discovered bank accounts

#### Scenario: List user connections endpoint
- **WHEN** `GET /api/v1/bank-connections` is called
- **THEN** it SHALL return all bank connections for the authenticated user

#### Scenario: Delete connection endpoint
- **WHEN** `DELETE /api/v1/bank-connections/{id}` is called
- **THEN** it SHALL delete the connection if owned by the authenticated user
- **AND** it SHALL respond with `204 No Content`

#### Scenario: Reconnect endpoint
- **WHEN** `POST /api/v1/bank-connections/{id}/reconnect` is called
- **THEN** it SHALL return a new auth link for the same institution

#### Scenario: All bank connection endpoints require authentication
- **WHEN** an unauthenticated request is made to any bank connection endpoint
- **THEN** it SHALL respond with `401 Unauthorized`

## REMOVED Requirements

### Requirement: GoCardless client interface
**Reason**: Replaced by Enable Banking client interface in `enable-banking-client` spec
**Migration**: Use `banksync.Client` interface with Enable Banking methods instead

### Requirement: GoCardless HTTP client implementation
**Reason**: Replaced by Enable Banking HTTP client implementation in `enable-banking-client` spec
**Migration**: Use `banksync.NewHTTPClient` with Enable Banking config (application ID + key path)

### Requirement: GoCardless config
**Reason**: Replaced by Enable Banking config in `enable-banking-client` spec
**Migration**: Replace `GOCARDLESS_*` env vars with `ENABLEBANKING_*` env vars
