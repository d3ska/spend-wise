## ADDED Requirements

### Requirement: Bank connection entity with GoCardless requisition tracking
The `model` package SHALL define a `BankConnection` entity with typed `BankConnectionID`, user ownership, GoCardless requisition reference, institution metadata, status tracking, and auth expiry.

#### Scenario: BankConnection fields
- **WHEN** a `BankConnection` struct is created
- **THEN** it SHALL have fields: `ID BankConnectionID`, `UserID UserID`, `InstitutionID string`, `InstitutionName string`, `RequisitionID string`, `Status BankConnectionStatus`, `AuthExpiresAt time.Time`, `CreatedAt time.Time`, `UpdatedAt time.Time`

#### Scenario: BankConnectionStatus enum values
- **WHEN** the `BankConnectionStatus` type is defined
- **THEN** it SHALL support values: `BankConnectionActive`, `BankConnectionExpired`, `BankConnectionRevoked`

### Requirement: Bank connection database migration
The database SHALL have a `bank_connections` table storing user-owned GoCardless requisition data.

#### Scenario: Bank connections table schema
- **WHEN** the bank connections migration is applied
- **THEN** the `bank_connections` table SHALL have columns: `id` (BIGSERIAL PK), `user_id` (BIGINT NOT NULL FK to users ON DELETE CASCADE), `institution_id` (TEXT NOT NULL), `institution_name` (TEXT NOT NULL), `requisition_id` (TEXT NOT NULL UNIQUE), `status` (TEXT NOT NULL DEFAULT 'active'), `auth_expires_at` (TIMESTAMPTZ NOT NULL), `created_at` (TIMESTAMPTZ NOT NULL DEFAULT NOW()), `updated_at` (TIMESTAMPTZ NOT NULL DEFAULT NOW())

#### Scenario: Migration rollback
- **WHEN** the bank connections migration is rolled back
- **THEN** the `bank_connections` table SHALL be dropped

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

#### Scenario: Delete bank connection
- **WHEN** `Delete(ctx, id, userID)` is called
- **THEN** it SHALL delete the connection only if owned by the given user
- **AND** associated `bank_accounts` and `workspace_bank_accounts` SHALL be removed via CASCADE

### Requirement: GoCardless client interface
The `banksync` package SHALL define a `Client` interface for interacting with the GoCardless Bank Account Data API v2.

#### Scenario: Client interface methods
- **WHEN** the `Client` interface is defined
- **THEN** it SHALL expose methods: `GetInstitutions(ctx, countryCode string) ([]Institution, error)`, `CreateAgreement(ctx, institutionID string) (Agreement, error)`, `CreateRequisition(ctx, institutionID, redirectURL, agreementID string) (Requisition, error)`, `GetRequisition(ctx, requisitionID string) (Requisition, error)`, `GetAccountTransactions(ctx, accountID string, dateFrom, dateTo time.Time) ([]Transaction, error)`

#### Scenario: Institution type fields
- **WHEN** an `Institution` struct is returned
- **THEN** it SHALL have fields: `ID string`, `Name string`, `Logo string`, `Countries []string`

#### Scenario: Requisition type fields
- **WHEN** a `Requisition` struct is returned
- **THEN** it SHALL have fields: `ID string`, `Link string`, `Status string`, `Accounts []string`

#### Scenario: Transaction type fields
- **WHEN** a GoCardless `Transaction` struct is returned
- **THEN** it SHALL have fields: `InternalTransactionID string`, `BookingDate string`, `TransactionAmount Amount`, `CreditorName string`, `DebtorName string`, `RemittanceInformation string`

#### Scenario: Amount type fields
- **WHEN** an `Amount` struct is returned
- **THEN** it SHALL have fields: `Amount string`, `Currency string`

### Requirement: GoCardless HTTP client implementation
The `banksync` package SHALL provide an HTTP client that authenticates with GoCardless using `secret_id`/`secret_key` and manages access/refresh tokens.

#### Scenario: Token creation
- **WHEN** the client is initialized
- **THEN** it SHALL call `POST /api/v2/token/new/` with `secret_id` and `secret_key` to obtain an access token and refresh token

#### Scenario: Token refresh
- **WHEN** an API call fails with 401
- **THEN** the client SHALL call `POST /api/v2/token/refresh/` with the refresh token and retry the original request once

#### Scenario: Institutions endpoint
- **WHEN** `GetInstitutions(ctx, "PL")` is called
- **THEN** the client SHALL call `GET /api/v2/institutions/?country=PL` and return the parsed institution list

#### Scenario: Transactions endpoint with date filtering
- **WHEN** `GetAccountTransactions(ctx, accountID, dateFrom, dateTo)` is called
- **THEN** the client SHALL call `GET /api/v2/accounts/{accountID}/transactions/?date_from={dateFrom}&date_to={dateTo}`
- **AND** it SHALL return only `booked` transactions (not `pending`)

### Requirement: Bank connection service with OAuth flow orchestration
The `service` package SHALL provide bank connection operations including the GoCardless requisition lifecycle.

#### Scenario: List available institutions
- **WHEN** `ListInstitutions(ctx, countryCode)` is called
- **THEN** it SHALL delegate to the GoCardless client and return the institution list

#### Scenario: Initiate bank connection
- **WHEN** `Initiate(ctx, userID, institutionID, redirectURL)` is called
- **THEN** it SHALL create a GoCardless agreement and requisition
- **AND** it SHALL persist a `BankConnection` with status `active` and `auth_expires_at` = now + 90 days
- **AND** it SHALL return the requisition link for user redirect

#### Scenario: Complete bank connection
- **WHEN** `Complete(ctx, userID, connectionID)` is called
- **THEN** it SHALL fetch the requisition from GoCardless to retrieve account IDs
- **AND** it SHALL create `BankAccount` records for each discovered account

#### Scenario: Delete bank connection requires ownership
- **WHEN** a user attempts to delete a connection they do not own
- **THEN** it SHALL return `model.ErrBankConnectionNotFound`

#### Scenario: Reconnect expired connection
- **WHEN** `Reconnect(ctx, userID, connectionID)` is called for an expired connection
- **THEN** it SHALL create a new GoCardless requisition for the same institution
- **AND** it SHALL update the connection's `requisition_id`, `status` to `active`, and `auth_expires_at` to now + 90 days

### Requirement: Bank connection HTTP handler
The `handler` package SHALL expose REST endpoints for bank connection management.

#### Scenario: List institutions endpoint
- **WHEN** `GET /api/v1/banks?country={code}` is called
- **THEN** it SHALL return the list of available banking institutions for that country
- **AND** the `country` query parameter SHALL default to `"PL"` if omitted

#### Scenario: Initiate connection endpoint
- **WHEN** `POST /api/v1/bank-connections` is called with `{"institution_id": "...", "redirect_url": "..."}`
- **THEN** it SHALL return the bank connection ID and the GoCardless auth link
- **AND** it SHALL respond with `201 Created`

#### Scenario: Complete connection endpoint
- **WHEN** `POST /api/v1/bank-connections/{id}/complete` is called
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

### Requirement: GoCardless config
The `config` package SHALL include GoCardless API credentials loaded from environment variables.

#### Scenario: GoCardless config fields
- **WHEN** configuration is loaded
- **THEN** it SHALL include `GoCardlessSecretID` (env: `GOCARDLESS_SECRET_ID`) and `GoCardlessSecretKey` (env: `GOCARDLESS_SECRET_KEY`)
- **AND** both fields SHALL default to empty strings

#### Scenario: GoCardless base URL config
- **WHEN** configuration is loaded
- **THEN** it SHALL include `GoCardlessBaseURL` (env: `GOCARDLESS_BASE_URL`) defaulting to `"https://bankaccountdata.gocardless.com"`
