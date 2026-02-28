## ADDED Requirements

### Requirement: User entity with OAuth provider identity
The `model` package SHALL define a `User` entity with typed `UserID`, `AuthProvider` enum (google|github), and fields: ID, Email, DisplayName, AvatarURL, Provider, ProviderID, CreatedAt, UpdatedAt. No password fields SHALL exist.

#### Scenario: User entity fields
- **WHEN** a `User` struct is created
- **THEN** it SHALL have fields: `ID UserID`, `Email string`, `DisplayName string`, `AvatarURL string`, `Provider AuthProvider`, `ProviderID string`, `CreatedAt time.Time`, `UpdatedAt time.Time`

#### Scenario: AuthProvider constants
- **WHEN** the `model` package is imported
- **THEN** `ProviderGoogle` SHALL equal `"google"` and `ProviderGitHub` SHALL equal `"github"`

#### Scenario: UserID typed wrapper
- **WHEN** a `UserID` is created from an `int64`
- **THEN** it SHALL be a distinct type (not a raw int64 alias)

### Requirement: Users table migration
The database SHALL have a `users` table with an `auth_provider` PostgreSQL enum type, unique email constraint, and provider+provider_id tracking.

#### Scenario: Users table schema
- **WHEN** migration 000001 is applied
- **THEN** the `users` table SHALL have columns: `id` (BIGSERIAL PK), `email` (TEXT UNIQUE NOT NULL), `display_name` (TEXT NOT NULL), `avatar_url` (TEXT NOT NULL DEFAULT ''), `provider` (auth_provider NOT NULL), `provider_id` (TEXT NOT NULL), `created_at` (TIMESTAMPTZ NOT NULL DEFAULT NOW()), `updated_at` (TIMESTAMPTZ NOT NULL DEFAULT NOW())

#### Scenario: Migration rollback
- **WHEN** migration 000001 is rolled back
- **THEN** the `users` table and `auth_provider` enum type SHALL be dropped

### Requirement: User store with create-or-find-by-email
The `store` package SHALL provide user persistence with `GetOrCreateByEmail` semantics: if a user with the given email exists, return it (with updated provider info); if not, create a new user.

#### Scenario: Create new user
- **WHEN** `GetOrCreateByEmail` is called with an email not in the database
- **THEN** a new user row SHALL be inserted and the created user SHALL be returned

#### Scenario: Find existing user by email
- **WHEN** `GetOrCreateByEmail` is called with an email already in the database
- **THEN** the existing user SHALL be returned with updated display_name, avatar_url, provider, and provider_id

#### Scenario: Get user by ID
- **WHEN** `GetByID(ctx, validUserID)` is called
- **THEN** the user with that ID SHALL be returned

#### Scenario: Get user by ID not found
- **WHEN** `GetByID(ctx, nonExistentID)` is called
- **THEN** it SHALL return `model.ErrUserNotFound`
