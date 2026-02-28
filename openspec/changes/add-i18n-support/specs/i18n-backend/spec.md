## ADDED Requirements

### Requirement: AppError type with machine-readable error codes
The `model` package SHALL define an `AppError` struct with a `Code` field (string, UPPER_SNAKE_CASE) and a `Message` field (string, English default). `AppError` SHALL implement the `error` interface. Each existing sentinel error in `model/errors.go` SHALL be replaced with (or wrapped by) an `AppError` value so that `errors.Is()` checks continue to work.

#### Scenario: AppError implements error interface
- **WHEN** an `AppError{Code: "WORKSPACE_NOT_FOUND", Message: "workspace not found"}` is created
- **THEN** its `.Error()` method SHALL return `"workspace not found"`

#### Scenario: AppError exposes code
- **WHEN** an `AppError` is used in error handling
- **THEN** the `Code` field SHALL be accessible as a string constant (e.g., `"WORKSPACE_NOT_FOUND"`)

#### Scenario: errors.Is compatibility preserved
- **WHEN** `errors.Is(err, model.ErrWorkspaceNotFound)` is called with an `AppError`-based error
- **THEN** it SHALL return `true`

### Requirement: Error code constants for all domain errors
The `model` package SHALL define a named constant for every error code. Each existing sentinel error SHALL have a corresponding code. Codes SHALL use UPPER_SNAKE_CASE format.

#### Scenario: Auth error codes
- **WHEN** auth-related errors occur
- **THEN** the following codes SHALL be available: `INVALID_CREDENTIALS`, `EMAIL_ALREADY_EXISTS`, `USER_NOT_FOUND`

#### Scenario: Workspace error codes
- **WHEN** workspace-related errors occur
- **THEN** the following codes SHALL be available: `WORKSPACE_NOT_FOUND`, `WORKSPACE_NAME_REQUIRED`, `WORKSPACE_NAME_TOO_LONG`, `WORKSPACE_DESCRIPTION_REQUIRED`, `WORKSPACE_DESCRIPTION_TOO_LONG`, `NOT_WORKSPACE_MEMBER`, `INSUFFICIENT_PERMISSION`, `CANNOT_REMOVE_SELF`

#### Scenario: Transaction error codes
- **WHEN** transaction-related errors occur
- **THEN** the following codes SHALL be available: `TRANSACTION_NOT_FOUND`, `TRANSACTION_NO_ENTRIES`, `ENTRY_SUM_MISMATCH`, `DUPLICATE_FINGERPRINT`, `TRANSACTION_AMOUNT_TOO_LARGE`, `TRANSACTION_DATE_TOO_FAR_IN_PAST`, `TRANSACTION_DATE_TOO_FAR_IN_FUTURE`

#### Scenario: Category error codes
- **WHEN** category-related errors occur
- **THEN** the following codes SHALL be available: `CATEGORY_NOT_FOUND`, `CATEGORY_NAME_REQUIRED`, `CATEGORY_NAME_TOO_LONG`, `CATEGORY_UNDELETABLE`

#### Scenario: Funding error codes
- **WHEN** funding-related errors occur
- **THEN** the following codes SHALL be available: `FUNDING_NOT_FOUND`, `CATEGORY_BUDGETS_EXCEED`

#### Scenario: Invite error codes
- **WHEN** invite-related errors occur
- **THEN** the following codes SHALL be available: `INVITE_NOT_FOUND`, `INVITE_EXPIRED`, `INVITE_USED`, `ALREADY_MEMBER`, `CANNOT_CHANGE_OWN_ROLE`, `INVALID_INVITE_ROLE`

#### Scenario: Rule error codes
- **WHEN** rule-related errors occur
- **THEN** the following codes SHALL be available: `RULE_NOT_FOUND`, `RULE_INVALID_PATTERN`, `RULE_PATTERN_TOO_LONG`

#### Scenario: Bank connection error codes
- **WHEN** bank-related errors occur
- **THEN** the following codes SHALL be available: `BANK_CONNECTION_NOT_FOUND`, `BANK_ACCOUNT_NOT_FOUND`, `BANK_ACCOUNT_ALREADY_EXISTS`, `BANK_CONNECTION_EXPIRED`

### Requirement: Coded error response helper
The `handler` package SHALL provide a `respondCodedError(w, status, code, message)` function that writes the new error envelope format. The existing `respondError` function SHALL be updated or replaced to use this format.

#### Scenario: Coded error response format
- **WHEN** `respondCodedError(w, 404, "WORKSPACE_NOT_FOUND", "workspace not found")` is called
- **THEN** the response body SHALL be `{"code":"WORKSPACE_NOT_FOUND","message":"workspace not found"}`
- **AND** the status code SHALL be `404`
- **AND** `Content-Type` SHALL be `application/json`

#### Scenario: Generic error without specific code
- **WHEN** an error occurs that does not map to a known code (e.g., handler-level validation like "invalid request body")
- **THEN** the handler SHALL use a generic code such as `INVALID_REQUEST` or `INTERNAL_ERROR` with the original message

### Requirement: handleServiceError extracts AppError codes
The `handleServiceError` function SHALL extract the `Code` field from `AppError` values and pass it to `respondCodedError`. The HTTP status mapping logic SHALL remain unchanged.

#### Scenario: Known AppError mapped to coded response
- **WHEN** `handleServiceError` receives `model.ErrWorkspaceNotFound`
- **THEN** it SHALL respond with status `404` and body `{"code":"WORKSPACE_NOT_FOUND","message":"workspace not found"}`

#### Scenario: Unknown error mapped to internal error
- **WHEN** `handleServiceError` receives an error that is not an `AppError`
- **THEN** it SHALL respond with status `500` and body `{"code":"INTERNAL_ERROR","message":"internal server error"}`

### Requirement: Handler inline errors use error codes
All inline `respondError` calls in handler files (e.g., "missing or invalid token", "invalid request body") SHALL be updated to use `respondCodedError` with appropriate codes.

#### Scenario: Authentication error uses code
- **WHEN** a request lacks a valid JWT token
- **THEN** the handler SHALL respond with `{"code":"UNAUTHORIZED","message":"missing or invalid token"}` and status `401`

#### Scenario: Invalid request body uses code
- **WHEN** a request body cannot be decoded
- **THEN** the handler SHALL respond with `{"code":"INVALID_REQUEST","message":"invalid request body"}` and status `400`

#### Scenario: Invalid path parameter uses code
- **WHEN** a workspace ID path parameter is not a valid integer
- **THEN** the handler SHALL respond with `{"code":"INVALID_REQUEST","message":"invalid workspace id"}` and status `400`
