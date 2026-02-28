## ADDED Requirements

### Requirement: Enable Banking client interface
The `banksync` package SHALL define a `Client` interface for interacting with the Enable Banking API.

#### Scenario: Client interface methods
- **WHEN** the `Client` interface is defined
- **THEN** it SHALL expose methods: `GetASPSPs(ctx, country string) ([]ASPSP, error)`, `StartAuth(ctx, aspspName, country, redirectURL string, validUntil time.Time) (AuthResult, error)`, `CreateSession(ctx, code string) (Session, error)`, `GetAccountTransactions(ctx, accountUID string, dateFrom, dateTo time.Time) ([]Transaction, error)`

#### Scenario: ASPSP type fields
- **WHEN** an `ASPSP` struct is returned
- **THEN** it SHALL have fields: `Name string`, `Country string`, `Logo string`, `BIC string`, `MaxConsentValidity int` (seconds)

#### Scenario: AuthResult type fields
- **WHEN** an `AuthResult` struct is returned
- **THEN** it SHALL have fields: `URL string` (bank redirect URL), `AuthorizationID string`

#### Scenario: Session type fields
- **WHEN** a `Session` struct is returned
- **THEN** it SHALL have fields: `SessionID string`, `Accounts []SessionAccount`

#### Scenario: SessionAccount type fields
- **WHEN** a `SessionAccount` struct is returned
- **THEN** it SHALL have fields: `UID string` (for API access), `AccountID string` (IBAN or other identifier), `Name string`, `Currency string`

#### Scenario: Transaction type fields
- **WHEN** an Enable Banking `Transaction` struct is returned
- **THEN** it SHALL have fields: `TransactionID string`, `BookingDate string`, `TransactionAmount Amount`, `CreditorName string`, `DebtorName string`, `RemittanceInformation []string`, `CreditDebitIndicator string`, `Status string`

#### Scenario: Amount type fields
- **WHEN** an `Amount` struct is returned
- **THEN** it SHALL have fields: `Amount string`, `Currency string`

### Requirement: JWT/RSA256 authentication
The Enable Banking HTTP client SHALL authenticate using self-signed JWT tokens with RSA256.

#### Scenario: JWT token generation
- **WHEN** the client needs to make an API request
- **THEN** it SHALL generate a JWT with header `{alg: "RS256", typ: "JWT", kid: applicationID}` and payload `{iss: "enablebanking.com", aud: "api.enablebanking.com", iat: now, exp: now + 3600}`

#### Scenario: RSA private key loading
- **WHEN** the client is initialized
- **THEN** it SHALL load an RSA private key from the PEM file path specified in config
- **AND** it SHALL support both PKCS1 (`RSA PRIVATE KEY`) and PKCS8 (`PRIVATE KEY`) PEM formats

#### Scenario: Authorization header
- **WHEN** making any API request
- **THEN** the client SHALL set the `Authorization` header to `Bearer {jwt}`

### Requirement: Enable Banking HTTP client implementation
The `banksync` package SHALL provide an HTTP client that implements the `Client` interface using the Enable Banking API.

#### Scenario: Get ASPSPs endpoint
- **WHEN** `GetASPSPs(ctx, "PL")` is called
- **THEN** the client SHALL call `GET https://api.enablebanking.com/aspsps?country=PL` and return the parsed ASPSP list

#### Scenario: Start auth endpoint
- **WHEN** `StartAuth(ctx, aspspName, country, redirectURL, validUntil)` is called
- **THEN** the client SHALL call `POST https://api.enablebanking.com/auth` with body `{"access": {"valid_until": validUntil}, "aspsp": {"name": aspspName, "country": country}, "state": uuid, "redirect_url": redirectURL, "psu_type": "personal"}`
- **AND** it SHALL return the `url` and `authorization_id` from the response

#### Scenario: Create session endpoint
- **WHEN** `CreateSession(ctx, code)` is called
- **THEN** the client SHALL call `POST https://api.enablebanking.com/sessions` with body `{"code": code}`
- **AND** it SHALL return the session ID and list of accounts with their UIDs and IBANs

#### Scenario: Get account transactions endpoint
- **WHEN** `GetAccountTransactions(ctx, accountUID, dateFrom, dateTo)` is called
- **THEN** the client SHALL call `GET https://api.enablebanking.com/accounts/{accountUID}/transactions?date_from={dateFrom}&date_to={dateTo}`
- **AND** it SHALL return only transactions with status `BOOK` (booked)

#### Scenario: Pagination via continuation_key
- **WHEN** a transactions response includes a `continuation_key`
- **THEN** the client SHALL make additional requests with the `continuation_key` parameter until all pages are fetched

### Requirement: Enable Banking config
The `config` package SHALL include Enable Banking API credentials loaded from environment variables.

#### Scenario: Enable Banking config fields
- **WHEN** configuration is loaded
- **THEN** it SHALL include `EnableBankingApplicationID` (env: `ENABLEBANKING_APP_ID`) and `EnableBankingKeyPath` (env: `ENABLEBANKING_KEY_PATH`)

#### Scenario: Enable Banking base URL config
- **WHEN** configuration is loaded
- **THEN** it SHALL include `EnableBankingBaseURL` (env: `ENABLEBANKING_BASE_URL`) defaulting to `"https://api.enablebanking.com"`
