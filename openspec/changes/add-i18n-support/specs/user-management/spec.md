## MODIFIED Requirements

### Requirement: User entity with OAuth provider identity
The `model` package SHALL define a `User` entity with typed `UserID`, `AuthProvider` enum (google|github), and fields: ID, Email, DisplayName, AvatarURL, Provider, ProviderID, PreferredLanguage, CreatedAt, UpdatedAt. No password fields SHALL exist.

#### Scenario: User entity fields
- **WHEN** a `User` struct is created
- **THEN** it SHALL have fields: `ID UserID`, `Email string`, `DisplayName string`, `AvatarURL string`, `Provider AuthProvider`, `ProviderID string`, `PreferredLanguage string`, `CreatedAt time.Time`, `UpdatedAt time.Time`

#### Scenario: PreferredLanguage default value
- **WHEN** a new `User` is created without specifying `PreferredLanguage`
- **THEN** `PreferredLanguage` SHALL default to `"en"`

#### Scenario: AuthProvider constants
- **WHEN** the `model` package is imported
- **THEN** `ProviderGoogle` SHALL equal `"google"` and `ProviderGitHub` SHALL equal `"github"`

#### Scenario: UserID typed wrapper
- **WHEN** a `UserID` is created from an `int64`
- **THEN** it SHALL be a distinct type (not a raw int64 alias)

### Requirement: Users table migration
The database SHALL have a `users` table with an `auth_provider` PostgreSQL enum type, unique email constraint, provider+provider_id tracking, and a `preferred_language` column.

#### Scenario: Users table schema
- **WHEN** migration 000001 is applied
- **THEN** the `users` table SHALL have columns: `id` (BIGSERIAL PK), `email` (TEXT UNIQUE NOT NULL), `display_name` (TEXT NOT NULL), `avatar_url` (TEXT NOT NULL DEFAULT ''), `provider` (auth_provider NOT NULL), `provider_id` (TEXT NOT NULL), `created_at` (TIMESTAMPTZ NOT NULL DEFAULT NOW()), `updated_at` (TIMESTAMPTZ NOT NULL DEFAULT NOW())

#### Scenario: preferred_language column added
- **WHEN** the i18n migration is applied
- **THEN** the `users` table SHALL have a `preferred_language` column of type `VARCHAR(5) NOT NULL DEFAULT 'en'`

#### Scenario: Migration rollback
- **WHEN** the i18n migration is rolled back
- **THEN** the `preferred_language` column SHALL be dropped from the `users` table

## ADDED Requirements

### Requirement: Update user preferred language
The store layer SHALL provide a method to update a user's `preferred_language` field. The service layer SHALL validate that the language is one of the supported values (`"en"`, `"pl"`).

#### Scenario: Update preferred language to Polish
- **WHEN** `UpdatePreferredLanguage(ctx, userID, "pl")` is called
- **THEN** the user's `preferred_language` column SHALL be set to `"pl"` and `updated_at` SHALL be refreshed

#### Scenario: Update with unsupported language
- **WHEN** `UpdatePreferredLanguage(ctx, userID, "fr")` is called
- **THEN** it SHALL return a validation error (unsupported language)

#### Scenario: Update for non-existent user
- **WHEN** `UpdatePreferredLanguage(ctx, nonExistentID, "pl")` is called
- **THEN** it SHALL return `model.ErrUserNotFound`

### Requirement: PATCH /api/v1/auth/me endpoint
The server SHALL expose `PATCH /api/v1/auth/me` (protected) that allows the authenticated user to update their profile fields. Initially, only `preferred_language` SHALL be supported.

#### Scenario: Update preferred language
- **WHEN** `PATCH /api/v1/auth/me` is called with `{"preferred_language": "pl"}`
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL contain the updated user profile including `"preferred_language": "pl"`

#### Scenario: Invalid language value
- **WHEN** `PATCH /api/v1/auth/me` is called with `{"preferred_language": "xx"}`
- **THEN** the response status SHALL be `400 Bad Request`
- **AND** the response body SHALL contain an error code indicating invalid language

#### Scenario: Unauthenticated request
- **WHEN** `PATCH /api/v1/auth/me` is called without a JWT
- **THEN** the response status SHALL be `401 Unauthorized`

#### Scenario: Empty body
- **WHEN** `PATCH /api/v1/auth/me` is called with `{}`
- **THEN** the response status SHALL be `200 OK` (no-op, returns current profile)

### Requirement: GET /api/v1/auth/me includes preferred_language
The `GET /api/v1/auth/me` response SHALL include the `preferred_language` field in the user profile JSON.

#### Scenario: Authenticated user response includes language
- **WHEN** `GET /api/v1/auth/me` is called with a valid JWT
- **THEN** the response body SHALL include `"preferred_language": "en"` (or the user's current preference)
