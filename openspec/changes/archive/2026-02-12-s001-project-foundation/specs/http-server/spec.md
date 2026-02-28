## ADDED Requirements

### Requirement: Chi-based HTTP server with configurable timeouts
The HTTP server SHALL use `go-chi/chi/v5` as the router and SHALL configure read/write timeouts from the application config.

#### Scenario: Server starts on configured address
- **WHEN** the server starts with `SERVER_HOST=0.0.0.0` and `SERVER_PORT=8080`
- **THEN** it SHALL listen on `0.0.0.0:8080`

#### Scenario: Timeouts are applied
- **WHEN** the server starts with `SERVER_READ_TIMEOUT=10s` and `SERVER_WRITE_TIMEOUT=30s`
- **THEN** `http.Server.ReadTimeout` SHALL be `10s` and `WriteTimeout` SHALL be `30s`

### Requirement: Health check endpoint
The server SHALL expose a `GET /healthz` endpoint that returns the server status without authentication.

#### Scenario: Health check returns OK
- **WHEN** `GET /healthz` is called
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL be `{"status":"ok"}`
- **AND** the `Content-Type` header SHALL be `application/json`

### Requirement: Graceful shutdown on signals
The server SHALL shut down gracefully when receiving `SIGINT` or `SIGTERM`, waiting up to 10 seconds for in-flight requests to complete.

#### Scenario: Graceful shutdown on SIGINT
- **WHEN** the server receives `SIGINT`
- **THEN** it SHALL stop accepting new connections
- **AND** it SHALL wait for in-flight requests to complete (up to 10 seconds)
- **AND** it SHALL exit with code 0

#### Scenario: Forced shutdown after timeout
- **WHEN** the server receives `SIGINT` and in-flight requests do not complete within 10 seconds
- **THEN** it SHALL force-close connections and exit

### Requirement: Structured logging with slog
The server SHALL use `log/slog` for all log output. Log messages SHALL include structured fields (key-value pairs).

#### Scenario: Server startup logged
- **WHEN** the server starts
- **THEN** it SHALL log `"server starting"` with the `addr` field

#### Scenario: Shutdown logged
- **WHEN** a shutdown signal is received
- **THEN** it SHALL log `"shutdown signal received"` with the `signal` field
- **AND** after shutdown completes, it SHALL log `"server stopped gracefully"`

### Requirement: JSON response helper
A `respondJSON` function SHALL write JSON responses with proper Content-Type header and status code.

#### Scenario: JSON response with data
- **WHEN** `respondJSON(w, 200, data)` is called
- **THEN** the response SHALL have `Content-Type: application/json`
- **AND** the status code SHALL be `200`
- **AND** the body SHALL be the JSON-encoded data

### Requirement: JSON error response helper
A `respondError` function SHALL write error responses in a consistent envelope format.

#### Scenario: Error response format
- **WHEN** `respondError(w, 400, "bad input")` is called
- **THEN** the response body SHALL be `{"error":"bad input"}`
- **AND** the status code SHALL be `400`

### Requirement: JSON request decoder with limits
A `decodeJSON` function SHALL decode request bodies with a 1MB size limit and reject unknown fields.

#### Scenario: Valid JSON decoded
- **WHEN** a request body contains valid JSON matching the target struct
- **THEN** `decodeJSON` SHALL populate the struct and return no error

#### Scenario: Body exceeds size limit
- **WHEN** a request body exceeds 1MB
- **THEN** `decodeJSON` SHALL return an error

#### Scenario: Unknown fields rejected
- **WHEN** a request body contains fields not in the target struct
- **THEN** `decodeJSON` SHALL return an error

### Requirement: Global middleware stack
The Chi router SHALL apply global middleware for all routes.

#### Scenario: Middleware applied
- **WHEN** any request is processed
- **THEN** the following Chi middleware SHALL be active: `RequestID`, `RealIP`, `Logger`, `Recoverer`

### Requirement: Entry point uses run() pattern
The `main()` function SHALL delegate to a `run()` function that returns an error. `main()` SHALL log the error and call `os.Exit(1)` on failure.

#### Scenario: Startup error handling
- **WHEN** `run()` returns an error (e.g., database connection failure)
- **THEN** `main()` SHALL log the error via `slog.Error` and exit with code 1
