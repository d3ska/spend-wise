## ADDED Requirements

### Requirement: CORS middleware for cross-origin frontend requests
The server SHALL apply CORS middleware globally using `go-chi/cors` that allows cross-origin requests from configured frontend origins with credentials support.

#### Scenario: Preflight request from allowed origin
- **WHEN** an `OPTIONS` request is received with `Origin: http://localhost:3000` and `CORS_ALLOWED_ORIGINS` includes `http://localhost:3000`
- **THEN** the response SHALL include `Access-Control-Allow-Origin: http://localhost:3000`
- **AND** the response SHALL include `Access-Control-Allow-Credentials: true`
- **AND** the response status SHALL be `200 OK`

#### Scenario: Actual request from allowed origin
- **WHEN** a `GET /api/v1/workspaces` request is received with `Origin: http://localhost:3000`
- **THEN** the response SHALL include `Access-Control-Allow-Origin: http://localhost:3000`
- **AND** the response SHALL include `Access-Control-Allow-Credentials: true`

#### Scenario: Request from disallowed origin
- **WHEN** a request is received with `Origin: http://evil.com` and that origin is not in `CORS_ALLOWED_ORIGINS`
- **THEN** the response SHALL NOT include `Access-Control-Allow-Origin` header

#### Scenario: Allowed HTTP methods
- **WHEN** a preflight `OPTIONS` request includes `Access-Control-Request-Method: DELETE`
- **THEN** the response `Access-Control-Allow-Methods` SHALL include GET, POST, PUT, DELETE, PATCH, and OPTIONS

#### Scenario: Allowed headers
- **WHEN** a preflight `OPTIONS` request includes `Access-Control-Request-Headers: Content-Type`
- **THEN** the response `Access-Control-Allow-Headers` SHALL include `Content-Type` and `Authorization`

### Requirement: CORS origins configuration from environment
The CORS allowed origins SHALL be configurable via the `CORS_ALLOWED_ORIGINS` environment variable as a comma-separated list.

#### Scenario: Default CORS origins
- **WHEN** `CORS_ALLOWED_ORIGINS` is not set
- **THEN** the allowed origins SHALL default to `http://localhost:3000`

#### Scenario: Multiple CORS origins
- **WHEN** `CORS_ALLOWED_ORIGINS=https://app.example.com,https://staging.example.com`
- **THEN** both origins SHALL be allowed

### Requirement: CORS middleware ordering
The CORS middleware SHALL be applied before the router's route matching so that preflight `OPTIONS` requests are handled before reaching auth middleware.

#### Scenario: Preflight does not hit auth
- **WHEN** a preflight `OPTIONS` request is sent to a protected route
- **THEN** the CORS middleware SHALL respond with CORS headers
- **AND** the request SHALL NOT reach the JWT authentication middleware
