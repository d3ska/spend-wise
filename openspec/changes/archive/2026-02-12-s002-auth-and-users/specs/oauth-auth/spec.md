## ADDED Requirements

### Requirement: Provider interface for OAuth2 authentication
The `auth` package SHALL define a `Provider` interface with methods `AuthCodeURL(state string) string`, `Exchange(ctx, code) (*oauth2.Token, error)`, and `FetchUser(ctx, token) (UserInfo, error)`. Each OAuth provider SHALL be a separate implementation.

#### Scenario: Google provider implements interface
- **WHEN** a Google provider is created with valid OAuth2 config
- **THEN** it SHALL implement the `Provider` interface
- **AND** `AuthCodeURL` SHALL return a Google OAuth2 authorization URL with the given state parameter

#### Scenario: GitHub provider implements interface
- **WHEN** a GitHub provider is created with valid OAuth2 config
- **THEN** it SHALL implement the `Provider` interface
- **AND** `AuthCodeURL` SHALL return a GitHub OAuth2 authorization URL with the given state parameter

### Requirement: OAuth2 code exchange with Google
The Google provider SHALL exchange an authorization code for an access token and fetch the user's profile from Google's userinfo API.

#### Scenario: Successful Google code exchange
- **WHEN** `Exchange(ctx, validCode)` is called on the Google provider
- **THEN** it SHALL return a valid `*oauth2.Token`

#### Scenario: Google user profile fetch
- **WHEN** `FetchUser(ctx, validToken)` is called on the Google provider
- **THEN** it SHALL return a `UserInfo` with email, display_name, avatar_url, and provider_id from Google's userinfo API

#### Scenario: Google exchange with invalid code
- **WHEN** `Exchange(ctx, invalidCode)` is called on the Google provider
- **THEN** it SHALL return an error

### Requirement: OAuth2 code exchange with GitHub
The GitHub provider SHALL exchange an authorization code for an access token and fetch the user's profile from GitHub's user API.

#### Scenario: Successful GitHub code exchange
- **WHEN** `Exchange(ctx, validCode)` is called on the GitHub provider
- **THEN** it SHALL return a valid `*oauth2.Token`

#### Scenario: GitHub user profile fetch
- **WHEN** `FetchUser(ctx, validToken)` is called on the GitHub provider
- **THEN** it SHALL return a `UserInfo` with email, display_name, avatar_url, and provider_id from GitHub's user API

#### Scenario: GitHub primary email fallback
- **WHEN** the GitHub user API returns no public email
- **THEN** the provider SHALL fetch from `GET /user/emails` and use the primary verified email

### Requirement: OAuth HTTP clients use explicit timeouts
All HTTP clients used for OAuth token exchange and profile fetching SHALL have a 10-second timeout. `http.DefaultClient` SHALL NOT be used.

#### Scenario: Timeout on slow provider
- **WHEN** a provider's token endpoint takes longer than 10 seconds
- **THEN** the request SHALL be cancelled with a timeout error

### Requirement: OAuth response bodies are closed
Every `http.Response.Body` obtained during OAuth profile fetching SHALL be closed via `defer resp.Body.Close()`.

#### Scenario: Response body closed after profile fetch
- **WHEN** `FetchUser` completes (success or failure after receiving a response)
- **THEN** the response body SHALL be closed

### Requirement: JWT token signing and verification
The `auth` package SHALL provide functions to sign and verify JWT tokens using HS256 with a configurable secret.

#### Scenario: Sign and verify round-trip
- **WHEN** a JWT is signed with `UserID=42` and secret `"test-secret"` with 24h expiry
- **THEN** verifying the token with the same secret SHALL return `UserID=42`

#### Scenario: Expired token rejected
- **WHEN** a JWT with an expired `exp` claim is verified
- **THEN** verification SHALL return an error

#### Scenario: Wrong secret rejected
- **WHEN** a JWT signed with secret `"A"` is verified with secret `"B"`
- **THEN** verification SHALL return an error

### Requirement: Authenticate flow (code -> JWT + user)
The `AuthService.Authenticate` method SHALL exchange an OAuth code for a user profile, create-or-find the user by email, and issue a JWT.

#### Scenario: First-time authentication creates user
- **WHEN** `Authenticate(ctx, "google", validCode)` is called for an email not in the database
- **THEN** a new user SHALL be created with the provider's email, display_name, avatar_url, provider, and provider_id
- **AND** a valid JWT SHALL be returned

#### Scenario: Returning user authentication
- **WHEN** `Authenticate(ctx, "google", validCode)` is called for an email already in the database
- **THEN** the existing user SHALL be returned (updated with latest provider info)
- **AND** a valid JWT SHALL be returned

#### Scenario: Cross-provider account linking
- **WHEN** a user first authenticates via Google with email `alice@example.com`
- **AND** later authenticates via GitHub with the same email `alice@example.com`
- **THEN** the same user record SHALL be returned both times

#### Scenario: Invalid provider name
- **WHEN** `Authenticate(ctx, "facebook", code)` is called with an unsupported provider
- **THEN** it SHALL return an error

### Requirement: CSRF state parameter
The `GET /api/v1/auth/{provider}/url` endpoint SHALL generate a random CSRF state token and include it in the OAuth authorization URL.

#### Scenario: State parameter in authorization URL
- **WHEN** `GET /api/v1/auth/google/url` is called
- **THEN** the response SHALL contain a `url` field with a `state` query parameter
- **AND** the response SHALL contain a `state` field matching the URL's state parameter
