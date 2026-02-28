## ADDED Requirements

### Requirement: Login page with OAuth provider buttons
The `/login` page SHALL display buttons for each supported OAuth provider (Google, GitHub). Each button SHALL initiate the OAuth flow.

#### Scenario: User clicks Google login
- **WHEN** the user clicks the "Sign in with Google" button
- **THEN** the app SHALL fetch `GET /api/v1/auth/google/url`
- **AND** redirect the browser to the returned `url`

#### Scenario: User clicks GitHub login
- **WHEN** the user clicks the "Sign in with GitHub" button
- **THEN** the app SHALL fetch `GET /api/v1/auth/github/url`
- **AND** redirect the browser to the returned `url`

### Requirement: OAuth callback handler
The `/auth/:provider/callback` route SHALL extract `code` and `state` from URL query parameters and exchange them with the backend.

#### Scenario: Successful OAuth callback
- **WHEN** the browser is redirected to `/auth/google/callback?code=abc&state=xyz`
- **THEN** the app SHALL POST `{code: "abc", state: "xyz"}` to `/api/v1/auth/google/callback`
- **AND** on success, redirect to `/dashboard`

#### Scenario: Failed OAuth callback
- **WHEN** the callback POST returns an error
- **THEN** the app SHALL redirect to `/login` with an error message displayed

### Requirement: Auth state management
The app SHALL determine auth state by calling `GET /api/v1/auth/me` on mount. The response provides the current user or a 401 if not authenticated.

#### Scenario: Authenticated user detected
- **WHEN** `GET /api/v1/auth/me` returns 200 with a user object
- **THEN** the user SHALL be stored in auth context and the app SHALL render protected routes

#### Scenario: Unauthenticated user detected
- **WHEN** `GET /api/v1/auth/me` returns 401
- **THEN** the app SHALL redirect to `/login`

### Requirement: Protected route guard
Protected routes SHALL be wrapped in an `AuthGuard` component that redirects unauthenticated users to `/login`.

#### Scenario: Unauthenticated access to protected route
- **WHEN** an unauthenticated user navigates to `/dashboard`
- **THEN** they SHALL be redirected to `/login`

#### Scenario: Authenticated access to protected route
- **WHEN** an authenticated user navigates to `/dashboard`
- **THEN** the dashboard page SHALL render normally

### Requirement: Logout functionality
The app SHALL provide a logout action that calls `POST /api/v1/auth/logout` and redirects to `/login`.

#### Scenario: User logs out
- **WHEN** the user clicks the logout button
- **THEN** the app SHALL call `POST /api/v1/auth/logout`
- **AND** clear the auth context
- **AND** redirect to `/login`
