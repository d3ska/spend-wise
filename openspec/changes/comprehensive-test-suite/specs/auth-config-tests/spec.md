## ADDED Requirements

### Requirement: Auth context round-trip tests
Tests SHALL verify that `withUserID` and `UserIDFromContext` correctly store and retrieve a user ID from a context. Tests SHALL also verify the "not present" case returns `(0, false)`.

#### Scenario: Store and retrieve user ID
- **WHEN** `withUserID(ctx, 42)` is called and `UserIDFromContext` is called on the resulting context
- **THEN** it returns `(42, true)`

#### Scenario: Missing user ID
- **WHEN** `UserIDFromContext` is called on a bare `context.Background()`
- **THEN** it returns `(0, false)`

#### Scenario: Different user IDs are independent
- **WHEN** `withUserID` is called twice with different IDs on nested contexts
- **THEN** `UserIDFromContext` returns the innermost ID

### Requirement: DevBypassMiddleware tests
Tests SHALL verify the dev bypass middleware: injects a user ID into context, returns 500 if user creation fails, caches the user across requests.

#### Scenario: Injects dev user into context
- **WHEN** a request passes through `DevBypassMiddleware` with a working user store
- **THEN** the next handler receives a context with a valid user ID

#### Scenario: User creation failure
- **WHEN** `DevBypassMiddleware` is used and the user store returns an error
- **THEN** the response status is 500 with a JSON error message

#### Scenario: Caches user across requests
- **WHEN** two consecutive requests pass through `DevBypassMiddleware`
- **THEN** `GetOrCreateByEmail` is called only once (via `sync.Once`)

### Requirement: ServerConfig.Addr tests
Tests SHALL verify `Addr()` formats the host and port correctly.

#### Scenario: Default values
- **WHEN** `Addr()` is called on `ServerConfig{Host: "0.0.0.0", Port: 8080}`
- **THEN** it returns `"0.0.0.0:8080"`

#### Scenario: Custom values
- **WHEN** `Addr()` is called on `ServerConfig{Host: "127.0.0.1", Port: 3000}`
- **THEN** it returns `"127.0.0.1:3000"`

### Requirement: DatabaseConfig.DSN tests
Tests SHALL verify `DSN()` formats the connection string correctly with all fields.

#### Scenario: Default config
- **WHEN** `DSN()` is called on a `DatabaseConfig` with default values
- **THEN** it returns `"postgres://postgres:postgres@localhost:5432/spendwise?sslmode=disable"`

#### Scenario: Custom config
- **WHEN** `DSN()` is called on a `DatabaseConfig` with custom Host, Port, User, Password, Name, SSLMode
- **THEN** the returned string contains all custom values in the correct positions

### Requirement: Config Load defaults test
Tests SHALL verify that `Load()` returns a Config with expected default values when no environment variables are set.

#### Scenario: Default server port
- **WHEN** `Load()` is called with no `SERVER_PORT` env var set
- **THEN** `Config.Server.Port` is 8080

#### Scenario: Default JWT secret
- **WHEN** `Load()` is called with no `JWT_SECRET` env var set
- **THEN** `Config.Auth.JWTSecret` is `"change-me-in-production"`

#### Scenario: Default database name
- **WHEN** `Load()` is called with no `DB_NAME` env var set
- **THEN** `Config.Database.Name` is `"spendwise"`
